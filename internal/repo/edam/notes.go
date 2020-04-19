package edam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rafaelespinoza/snbackfill/internal/entity"

	"github.com/dreampuf/evernote-sdk-golang/edam"
)

// Notes handles input/output for notes from the Evernote EDAM API.
type Notes struct {
	rqp NotesRemoteQueryParams
}

// NewNotesRepo constructs a Notes repository.
func NewNotesRepo(rqp *NotesRemoteQueryParams) (entity.LocalRemoteRepo, error) {
	if rqp == nil {
		rqp = &NotesRemoteQueryParams{TagIDs: make([]string, 0)}
	}
	return &Notes{rqp: *rqp}, nil
}

// FetchRemote gets Notes from the Evernote EDAM API. It can automatically
// perform pagination based on the param argument which should be of the
// concrete type, NotesQuery. Multiple API calls will be made until there are no
// more remaining results.
func (n *Notes) FetchRemote(ctx context.Context) (out []entity.LinkID, err error) {
	var (
		s *store
	)
	if s, err = initStore(ctx); err != nil {
		return
	}
	filter := n.rqp.toFilter()
	pageSize := int32(n.rqp.PageSize)
	pagination := newPaginator(int32(n.rqp.LoIndex), int32(n.rqp.HiIndex))

	yes := true
	// resultSpec tells evernote which fields to include in the search. By
	// default, only the note GUID is returned. To include more note fields, you
	// must specify the option with a pointer to true.
	resultSpec := &edam.NotesMetadataResultSpec{
		IncludeAttributes:   &yes,
		IncludeCreated:      &yes,
		IncludeNotebookGuid: &yes,
		IncludeTagGuids:     &yes,
		IncludeTitle:        &yes,
		IncludeUpdated:      &yes,
	}
	out = make([]entity.LinkID, 0)

	for !pagination.done {
		var ierr error
		if n.rqp.Verbose {
			fmt.Printf("fetching metadata, running total: %d\n", len(out))
		}
		notesMetadataList, ierr := s.noteClient.FindNotesMetadata(
			ctx,
			s.noteClient.token,
			filter,
			pagination.currOffset,
			pageSize,
			resultSpec,
		)
		if ierr != nil {
			err = makeError(ierr)
			return
		}
		notesMetadata := notesMetadataList.GetNotes()
		numResultsInRange := len(notesMetadata)
		numTotalResults := notesMetadataList.GetTotalNotes()
		subList := make([]entity.LinkID, numResultsInRange)
		ierr = pagination.update(
			notesMetadataList.GetStartIndex(),
			int32(numResultsInRange),
		)
		if ierr != nil {
			err = ierr
			return
		}

		resultSpec := &edam.NoteResultSpec{IncludeContent: &yes}
		if n.rqp.Verbose {
			fmt.Printf("\tdone fetching metadata, results %d\n", len(notesMetadata))
		}
		for i, noteMeta := range notesMetadata {
			noteID := noteMeta.GetGUID()
			if n.rqp.Verbose && numResultsInRange > 1 && i%(numResultsInRange/2) == 0 {
				fmt.Printf("\tfetching note content %d/%d\n", len(out)+i, numTotalResults)
			}
			result, ierr := s.noteClient.GetNoteWithResultSpec(
				ctx,
				s.noteClient.token,
				noteID,
				resultSpec,
			)
			if ierr != nil {
				err = fmt.Errorf(
					"%w, noteID: %q, noteContentLength %d",
					ierr, noteID, noteMeta.GetContentLength(),
				)
				return
			}
			note, ierr := newNote(noteMeta, result.GetContent())
			if ierr != nil {
				err = ierr
				return
			}

			subList[i] = note
		}
		out = append(out, subList...)
		if n.rqp.Verbose {
			fmt.Printf("\tdone fetching contents, total so far: %d\n", len(out))
		}
	}
	return
}

// NotesRemoteQueryParams is a set of named options for listing Evernote notes.
type NotesRemoteQueryParams struct {
	LoIndex    int
	HiIndex    int
	PageSize   int
	TagIDs     []string
	NotebookID string
	Verbose    bool
}

// toFilter converts the params note search options. The default order is by
// created at, which is good because it doesn't change.
func (p *NotesRemoteQueryParams) toFilter() *edam.NoteFilter {
	order := int32(edam.NoteSortOrder_CREATED)
	ascending := true
	filt := &edam.NoteFilter{
		Order:     &order,
		Ascending: &ascending,
	}
	if p.TagIDs != nil && len(p.TagIDs) > 0 {
		tagGUIDs := make([]edam.GUID, len(p.TagIDs))
		for i, id := range p.TagIDs {
			tagGUIDs[i] = edam.GUID(id)
		}
		filt.TagGuids = tagGUIDs
	}
	if p.NotebookID != "" {
		guid := edam.GUID(p.NotebookID)
		filt.NotebookGuid = &guid
	}
	return filt
}

var errPaginationOrdering = errors.New("pagination ordering probably messed up")

// A paginator helps manage pagination state.
type paginator struct {
	lo         int32
	hi         int32
	currOffset int32
	done       bool
}

func newPaginator(lo, hi int32) *paginator {
	if hi < 0 {
		hi = 1<<31 - 1
	}
	return &paginator{lo: lo, currOffset: lo, hi: hi}
}

func (p *paginator) update(startIndex, totalResults int32) (err error) {
	if totalResults < 1 {
		p.done = true
		return
	}
	// juuuust in case it's negative, update this field after checking the sign.
	p.currOffset += totalResults
	if p.currOffset > p.hi {
		p.done = true
		return
	}
	if p.currOffset != startIndex+totalResults {
		err = errPaginationOrdering
		return
	}
	return
}

func newNote(noteMeta *edam.NoteMetadata, content string) (resource entity.LinkID, err error) {
	// cannot fill the Tags field (name of the tag itself) from here, but it can
	// be "backfilled" after grabbing the tag data in a separate request.
	id := string(noteMeta.GetGUID())
	note := entity.Note{
		ID:         id,
		Title:      noteMeta.GetTitle(),
		NotebookID: noteMeta.GetNotebookGuid(),
		Content:    content,
		CreatedAt:  makeTimestamp(noteMeta.GetCreated()),
		UpdatedAt:  makeTimestamp(noteMeta.GetUpdated()),
	}
	tagIDs := noteMeta.GetTagGuids()
	note.TagIDs = make([]string, len(tagIDs))
	for j, tagID := range tagIDs {
		note.TagIDs[j] = string(tagID)
	}
	attrs := noteMeta.GetAttributes()
	note.Attributes = &entity.Attributes{
		ContentClass:      attrs.GetContentClass(),
		Source:            attrs.GetSource(),
		SourceApplication: attrs.GetSourceApplication(),
		SourceURL:         attrs.GetSourceURL(),
	}
	resource = &Note{
		Note:      &note,
		ServiceID: &entity.ServiceID{Value: id},
	}
	return
}

// ReadLocal reads and parses notes saved in a local JSON file.
func (n *Notes) ReadLocal(ctx context.Context, r io.Reader) (out []entity.LinkID, err error) {
	decoder := json.NewDecoder(r)
	var resources []*Note
	if err = decoder.Decode(&resources); err != nil {
		return
	}
	out = make([]entity.LinkID, len(resources))
	for i, res := range resources {
		res.ServiceID = &entity.ServiceID{Value: res.ID}
		out[i] = res
	}
	return
}

// Note represents a note in an Evernote EDAM API call. It also provides methods
// to match with a Note in StandardNotes.
type Note struct {
	*entity.Note
	*entity.ServiceID
}

func (n *Note) LinkValues() []string {
	return []string{
		fmtTime(n.CreatedAt),
		n.Title,
		fmtTime(n.UpdatedAt),
	}
}
