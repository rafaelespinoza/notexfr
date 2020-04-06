package interactor

import (
	"context"
	"fmt"
	"time"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/repo"
	"github.com/rafaelespinoza/snbackfill/internal/repo/edam"
	"github.com/rafaelespinoza/snbackfill/internal/repo/sn"
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

func BackfillSN(ctx context.Context, opts *BackfillOpts) (out []lib.LinkID, err error) {
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
	enNoteDegrees := make([]map[string][]lib.LinkID, numNoteLinks)
	for i := 0; i < numNoteLinks; i++ {
		enNoteDegrees[i] = make(map[string][]lib.LinkID)
	}
	err = evernote.notes.each(func(note lib.LinkID) (ierr error) {
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
				enNoteDegrees[i][link] = []lib.LinkID{note}
			} else {
				enNoteDegrees[i][link] = append(list, note)
			}
		}

		return
	})
	if err != nil {
		return
	}
	var notes []lib.LinkID
	err = standardnotes.notes.each(func(note lib.LinkID) (ierr error) {
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
		newRepo  func() (lib.RepoLocal, error)
		filename string
		target   *keyedItems
	}{
		{
			newRepo: func() (out lib.RepoLocal, err error) {
				out, err = edam.NewNotebooksRepo()
				return
			},
			filename: opts.EvernoteFilenames.Notebooks,
			target:   &out.notebooks,
		},
		{
			newRepo: func() (out lib.RepoLocal, err error) {
				out, err = edam.NewNotesRepo(nil)
				return
			},
			filename: opts.EvernoteFilenames.Notes,
			target:   &out.notes,
		},
		{
			newRepo: func() (out lib.RepoLocal, err error) {
				out, err = edam.NewTagsRepo()
				return
			},
			filename: opts.EvernoteFilenames.Tags,
			target:   &out.tags,
		},
	}

	var (
		repository lib.RepoLocal
		list       []lib.LinkID
	)

	for _, input := range inputs {
		if repository, err = input.newRepo(); err != nil {
			return
		}
		if list, err = readLocalFile(ctx, repository, input.filename); err != nil {
			return
		}
		collection := make(keyedItems)
		for _, item := range list {
			collection[item.GetID()] = item
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

	noteItems, tagItems := make(map[string]lib.LinkID), make(map[string]lib.LinkID)
	for _, item := range notes {
		noteItems[item.GetID()] = item
	}
	for _, item := range tags {
		tagItems[item.GetID()] = item
	}

	out = &serviceItems{
		notes: noteItems,
		tags:  tagItems,
	}
	return
}

// keyedItems is a collection resources that is indexed by some unique key,
// such as an ID, where all items originate from the same service provider.
type keyedItems map[string]lib.LinkID

func (m keyedItems) each(cb func(item lib.LinkID) error) (err error) {
	for key, item := range m {
		if err = cb(item); err != nil {
			err = fmt.Errorf("%w; key: %q", err, key)
			return
		}
	}
	return
}

// MatchTags attempts to reconcile tags in Evernote and StandardNotes by reading
// metadata in local files and comparing values.
func MatchTags(ctx context.Context, opts *BackfillOpts) (tags []lib.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository lib.LocalRemoteRepo
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
func MatchNotes(ctx context.Context, opts *BackfillOpts) (notes []lib.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository lib.LocalRemoteRepo
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
func ReconcileNotebooks(ctx context.Context, opts *BackfillOpts) (notebookTags []lib.LinkID, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var repository lib.LocalRemoteRepo
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
func (b *BackfillOpts) readParseFiles(ctx context.Context, enRepo lib.RepoLocal, enFilename string) (enOut, snNotes, snTags []lib.LinkID, err error) {
	enOut, err = readLocalFile(ctx, enRepo, enFilename)
	if err != nil {
		return
	}
	snNotes, snTags, err = sn.ReadConversionFile(b.StandardNotesFilename)
	return
}

// link associates resources among services by bucketing and matching based on
// previously agreed upon field values.
func (b *BackfillOpts) link(enResources, snResources []lib.LinkID) (out []lib.LinkID, err error) {
	var (
		links []string
		list  []lib.LinkID
		ok    bool
	)

	groupedEN := make(map[string][]lib.LinkID)
	for i, resource := range enResources {
		links = resource.LinkValues()
		if len(links) < 1 {
			err = fmt.Errorf("%w; evernote [%d]", errLinksEmpty, i)
			return
		}
		if list, ok = groupedEN[links[0]]; !ok {
			groupedEN[links[0]] = []lib.LinkID{resource}
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
func (b *BackfillOpts) makeNotebooks(enNotebooks []lib.LinkID, prefix string, snTags []lib.LinkID) (out []lib.LinkID, err error) {
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
	lib.LinkID `json:"Item"`
	EvernoteID lib.Resource
}
