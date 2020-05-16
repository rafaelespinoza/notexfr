package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelespinoza/notexfr/internal/entity"
	"github.com/rafaelespinoza/notexfr/internal/repo"
	"github.com/rafaelespinoza/notexfr/internal/repo/edam"
	"github.com/rafaelespinoza/notexfr/internal/repo/sn"
)

// BackfillParams is a set of named parameters for performing a backfill on
// Evernote, StandardNotes data.
type BackfillParams struct {
	EvernoteFilenames     struct{ Notebooks, Notes, Tags string }
	StandardNotesFilename string
	OutputFilenames       struct{ Notebooks, Notes, Tags string }
	Verbose               bool
}

func BackfillSN(ctx context.Context, opts *BackfillParams) (out []entity.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var evernote, standardnotes *serviceItems

	if evernote, err = initEvernoteItems(ctx, opts); err != nil {
		return
	}
	if standardnotes, err = initStandardNotesItems(ctx, opts); err != nil {
		return
	}
	const numNoteLinks = 3
	enNoteDegrees := make([]map[string][]entity.LinkID, numNoteLinks)
	for i := 0; i < numNoteLinks; i++ {
		enNoteDegrees[i] = make(map[string][]entity.LinkID)
	}
	err = evernote.notes.each(func(note entity.LinkID) (ierr error) {
		links := note.LinkValues()
		if len(links) != numNoteLinks {
			ierr = fmt.Errorf(
				"expected links length to be %d; got %d; evernote %q",
				numNoteLinks, len(links), note.GetID(),
			)
			return
		}

		for i, link := range links {
			if list, ok := enNoteDegrees[i][link]; !ok {
				enNoteDegrees[i][link] = []entity.LinkID{note}
			} else {
				enNoteDegrees[i][link] = append(list, note)
			}
		}

		return
	})
	if err != nil {
		return
	}
	var notes []entity.LinkID
	err = standardnotes.notes.each(func(note entity.LinkID) (ierr error) {
		links := note.LinkValues()
		if len(links) != numNoteLinks {
			ierr = fmt.Errorf(
				"expected links length to be %d; got %d; standardnotes %q",
				numNoteLinks, len(links), note.GetID(),
			)
			return
		}
		for i, link := range links {
			enNotes, ok := enNoteDegrees[i][link]
			if !ok {
				// This is an anomaly. TODO: log or something.
				continue
			}
			if len(enNotes) == 1 && link == enNotes[0].LinkValues()[i] {
				enNote := enNotes[0].(*edam.Note)
				snNote := note.(*sn.Note)
				snNote.AppendTags(enNote.NotebookID)
				notes = append(notes, &FromENToSN{
					LinkID:     snNote,
					EvernoteID: repo.NewServiceID(enNote.ID),
				})
				break
			}
		}
		return
	})
	if err != nil {
		return
	}
	if err = writeResources(notes, opts.OutputFilenames.Notes, opts.Verbose, "backfilled notes"); err != nil {
		return
	}
	out = notes
	return
}

// serviceItems manages items from one service.
type serviceItems struct {
	notebooks, notes, tags keyedItems
}

func initEvernoteItems(ctx context.Context, opts *BackfillParams) (out *serviceItems, err error) {
	out = &serviceItems{}
	inputs := []struct {
		newRepo  func() (entity.RepoLocal, error)
		filename string
		target   *keyedItems
	}{
		{
			newRepo: func() (out entity.RepoLocal, err error) {
				out, err = edam.NewNotebooksRepo()
				return
			},
			filename: opts.EvernoteFilenames.Notebooks,
			target:   &out.notebooks,
		},
		{
			newRepo: func() (out entity.RepoLocal, err error) {
				out, err = edam.NewNotesRepo(nil)
				return
			},
			filename: opts.EvernoteFilenames.Notes,
			target:   &out.notes,
		},
		{
			newRepo: func() (out entity.RepoLocal, err error) {
				out, err = edam.NewTagsRepo()
				return
			},
			filename: opts.EvernoteFilenames.Tags,
			target:   &out.tags,
		},
	}

	var (
		repository entity.RepoLocal
		list       []entity.LinkID
	)

	for _, input := range inputs {
		if repository, err = input.newRepo(); err != nil {
			return
		}
		if list, err = readLocalFile(ctx, repository, input.filename); err != nil {
			return
		}
		collection := makeKeyedItems(len(list))
		for i, item := range list {
			itemID := item.GetID()
			collection.items[itemID] = item
			collection.keys[i] = itemID
		}
		// this might lead to memory leaks. check it.
		*input.target = collection
	}
	return
}

func initStandardNotesItems(ctx context.Context, opts *BackfillParams) (out *serviceItems, err error) {
	notes, tags, err := sn.ReadConversionFile(opts.StandardNotesFilename)
	if err != nil {
		return
	}

	keyedNoteItems, keyedTagItems := makeKeyedItems(len(notes)), makeKeyedItems(len(tags))
	for i, item := range notes {
		itemID := item.GetID()
		keyedNoteItems.items[itemID] = item
		keyedNoteItems.keys[i] = itemID
	}
	for i, item := range tags {
		itemID := item.GetID()
		keyedTagItems.items[itemID] = item
		keyedTagItems.keys[i] = itemID
	}

	out = &serviceItems{
		notes: keyedNoteItems,
		tags:  keyedTagItems,
	}
	return
}

// keyedItems is a collection resources that is indexed by some unique key,
// such as an ID, where all items originate from the same service provider.
// The keys field preserves the insertion order so you can iterate through items
// for easier comparison to the original inputs while still providing the
// benefit of constant-time lookups.
type keyedItems struct {
	keys  []string
	items map[string]entity.LinkID
}

func makeKeyedItems(numKeys int) keyedItems {
	return keyedItems{
		items: make(map[string]entity.LinkID),
		keys:  make([]string, numKeys),
	}
}

func (m keyedItems) each(cb func(item entity.LinkID) error) (err error) {
	for i, key := range m.keys {
		if err = cb(m.items[key]); err != nil {
			err = fmt.Errorf("%w; ind: %d, key: %q", err, i, key)
			return
		}
	}
	return
}

// FromENToSN is a StandardNotes resource that has a relationship to an Evernote
// resource. The EvernoteID field can just be a ServiceID, it does not need to
// be an entire Evernote resource.
type FromENToSN struct {
	entity.LinkID `json:"Item"`
	EvernoteID    entity.Resource
}
