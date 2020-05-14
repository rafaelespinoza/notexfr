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

var (
	errLinksEmpty    = fmt.Errorf("links empty")
	errTypeAssertion = fmt.Errorf("type assertion error")
)

// BackfillOpts is a set of named options for performing reconciliation on
// resources between Evernote and StandardNotes.
type BackfillOpts struct {
	EvernoteFilenames     struct{ Notebooks, Notes, Tags string }
	StandardNotesFilename string
	OutputFilenames       struct{ Notebooks, Notes, Tags string }
	Verbose               bool
}

func BackfillSN(ctx context.Context, opts *BackfillOpts) (out []entity.LinkID, err error) {
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

func initEvernoteItems(ctx context.Context, opts *BackfillOpts) (out *serviceItems, err error) {
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

func initStandardNotesItems(ctx context.Context, opts *BackfillOpts) (out *serviceItems, err error) {
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

// MatchTags attempts to reconcile tags in Evernote and StandardNotes by reading
// metadata in local files and comparing values.
func MatchTags(ctx context.Context, opts *BackfillOpts) (tags []entity.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository entity.LocalRemoteRepo
	if repository, err = edam.NewTagsRepo(); err != nil {
		return
	}
	enTags, _, snTags, err := opts.readParseFiles(ctx, repository, opts.EvernoteFilenames.Tags)
	if err != nil {
		return
	}

	tags, err = opts.link(enTags, snTags)
	if err != nil {
		return
	}
	err = writeResources(tags, opts.OutputFilenames.Tags, opts.Verbose, "matched tags")
	return
}

// MatchNotes attempts to reconcile notes in Evernote and StandardNotes by
// reading metadata in local files and comparing values.
func MatchNotes(ctx context.Context, opts *BackfillOpts) (notes []entity.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository entity.LocalRemoteRepo
	if repository, err = edam.NewNotesRepo(nil); err != nil {
		return
	}
	enNotes, snNotes, _, err := opts.readParseFiles(ctx, repository, opts.EvernoteFilenames.Notes)
	if err != nil {
		return
	}

	// Match on evernote.CreatedAt == standardnotes.created_at.
	// Set the ENID, SNID fields on a match.
	// If there's more than 1 possible match, then further refine the match by
	// comparing Title, UpdatedAt.
	notes, err = opts.link(enNotes, snNotes)
	if err != nil {
		return
	}
	err = writeResources(notes, opts.OutputFilenames.Notes, opts.Verbose, "matched notes")
	return
}

// ReconcileNotebooks approximates Notebooks in StandardNotes, which on its own,
// does not have the concept of Notebooks; instead it uses Tags. This function
// uses Evernote Notebook data to create StandardNotes Tag data.
func ReconcileNotebooks(ctx context.Context, opts *BackfillOpts) (notebookTags []entity.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository entity.LocalRemoteRepo
	if repository, err = edam.NewNotebooksRepo(); err != nil {
		return
	}
	enNotebooks, _, snTags, err := opts.readParseFiles(ctx, repository, opts.EvernoteFilenames.Notebooks)
	if err != nil {
		return
	}

	notebookTags, err = opts.makeNotebooks(enNotebooks, "conflict - ", snTags)
	if err != nil {
		return
	}
	err = writeResources(notebookTags, opts.OutputFilenames.Notebooks, opts.Verbose, "reconciled notebooks")
	return
}

// readParseFiles returns evernote resources as enOut based on the concrete type
// of enRepo. The standardnotes resources, snNotes and snTags, will always be a
// notes and tags respectively. Callers can use a blank identifier if one of
// the resources isn't needed.
func (b *BackfillOpts) readParseFiles(ctx context.Context, enRepo entity.RepoLocal, enFilename string) (enOut, snNotes, snTags []entity.LinkID, err error) {
	enOut, err = readLocalFile(ctx, enRepo, enFilename)
	if err != nil {
		return
	}
	snNotes, snTags, err = sn.ReadConversionFile(b.StandardNotesFilename)
	return
}

// link associates resources among services by bucketing and matching based on
// previously agreed upon field values.
func (b *BackfillOpts) link(enResources, snResources []entity.LinkID) (out []entity.LinkID, err error) {
	var (
		links []string
		list  []entity.LinkID
		ok    bool
	)

	groupedEN := make(map[string][]entity.LinkID)
	for i, resource := range enResources {
		links = resource.LinkValues()
		if len(links) < 1 {
			err = fmt.Errorf("%w; evernote [%d]", errLinksEmpty, i)
			return
		}
		if list, ok = groupedEN[links[0]]; !ok {
			groupedEN[links[0]] = []entity.LinkID{resource}
		} else {
			groupedEN[links[0]] = append(list, resource)
		}
	}

	var conv *FromENToSN

	for i, resource := range snResources {
		links = resource.LinkValues()
		if len(links) < 1 {
			err = fmt.Errorf("%w; standardnotes [%d]", errLinksEmpty, i)
			return
		}
		if list, ok = groupedEN[links[0]]; !ok {
			// TODO: log or something
			continue
		}
		// TODO: handle list with length > 1
		conv = &FromENToSN{
			LinkID:     resource,
			EvernoteID: repo.NewServiceID(list[0].GetID()),
		}
		out = append(out, conv)
	}
	return
}

// makeNotebooks converts Evernote notebooks into StandardNotes Tags. We know
// that there are no Evernote Notebooks in the export data that's used to create
// the initial StandardNotes import data and that StandardNotes does not have
// Notebooks.
func (b *BackfillOpts) makeNotebooks(enNotebooks []entity.LinkID, prefix string, snTags []entity.LinkID) (out []entity.LinkID, err error) {
	allNotebooks := make(map[string]*edam.Notebook)

	for _, resource := range enNotebooks {
		var ok bool
		notebook, ok := resource.(*edam.Notebook)
		if !ok {
			return nil, fmt.Errorf(
				"%w; expected input to be concrete type %T",
				errTypeAssertion, &edam.Notebook{},
			)
		}
		// Evernote Notebooks cannot have the same case-insensitive name, so
		// this is a safe assumption.
		allNotebooks[notebook.Name] = notebook

		if notebook.Stack == "" {
			continue
		}
		if _, ok = allNotebooks[notebook.Stack]; ok {
			continue
		}

		// This notebook does not exist in either Evernote or StandardNotes.
		// It's meant to be metadata, an approximation of an Evernote Notebook's
		// stack for use in StandardNotes.
		// TODO: handle naming conflicts between Notebook.Stack, Notebook.Name
		allNotebooks[notebook.Stack] = edam.NewNotebook(notebook.Stack)
	}

	// Each imported Evernote Tag will have a corresponding StandardNotes Tag,
	// but there might be a naming overlap since the uniqueness constraint in
	// Evernote does not span between Notebook and Tag resources. If there is an
	// overlap, then prefix the name of the new resource.
	for _, resource := range snTags {
		var ok bool
		tag, ok := resource.(*sn.Tag)
		if !ok {
			return nil, fmt.Errorf(
				"%w; expected input to be concrete type %T",
				errTypeAssertion, &sn.Tag{},
			)
		}
		title := tag.Content.Title
		if nb, ok := allNotebooks[title]; ok {
			nb.Notebook.Name = prefix + title
			allNotebooks[prefix+title] = nb
		}
	}

	// the output order will be random-ish
	for _, nb := range allNotebooks {
		out = append(out, &FromENToSN{
			LinkID: sn.NewTag(
				nb.Notebook.Name,
				nb.CreatedAt,
				nb.UpdatedAt,
			),
			EvernoteID: repo.NewServiceID(nb.GetID()),
		})
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
