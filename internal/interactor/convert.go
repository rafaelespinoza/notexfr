package interactor

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelespinoza/snbackfill/internal/entity"
	"github.com/rafaelespinoza/snbackfill/internal/repo/edam"
	"github.com/rafaelespinoza/snbackfill/internal/repo/enex"
	"github.com/rafaelespinoza/snbackfill/internal/repo/sn"
)

// ConvertOptions are named inputs and outputs for converting data between
// different formats.
type ConvertOptions struct {
	InputFilenames                struct{ Notebooks, Notes, Tags string }
	InputFilename, OutputFilename string
}

// SN is the output of converting resources to the import, export format for
// StandardNotes. The item schema is described at:
// https://docs.standardnotes.org/specification/sync/#items
type SN struct {
	Items []entity.LinkID `json:"items"`
}

// ConvertEDAMToStandardNotes replicates the existing data conversion tools at
// https://dashboard.standardnotes.org/tools.
func ConvertEDAMToStandardNotes(ctx context.Context, opts ConvertOptions) (out *SN, err error) {
	evernote, err := initEvernoteItems(ctx, &BackfillOpts{
		EvernoteFilenames: struct{ Notebooks, Notes, Tags string }{
			Notebooks: opts.InputFilenames.Notebooks,
			Notes:     opts.InputFilenames.Notes,
			Tags:      opts.InputFilenames.Tags,
		},
	})
	if err != nil {
		return
	}

	var combinedEnexResources []entity.LinkID
	for _, source := range []keyedItems{evernote.notes, evernote.tags, evernote.notebooks} {
		source.each(func(item entity.LinkID) error {
			combinedEnexResources = append(combinedEnexResources, item)
			return nil
		})
	}

	converter := &edamToSN{snConverter: snConverter{}}
	items, err := converter.convertToSN(combinedEnexResources)
	if err != nil {
		return
	}
	out = &SN{items}
	err = writeResources(
		out,
		opts.OutputFilename,
		false,
		"standardnotes resources",
	)
	return
}

// ConvertENEXToStandardNotes replicates the existing data conversion tools at
// https://dashboard.standardnotes.org/tools.
func ConvertENEXToStandardNotes(ctx context.Context, opts ConvertOptions) (out *SN, err error) {
	var (
		repository                   entity.RepoLocal
		converter                    *enexToSN
		origResources, convResources []entity.LinkID
	)

	if repository, err = enex.NewFileRepo(); err != nil {
		return
	}
	origResources, err = readLocalFile(ctx, repository, opts.InputFilename)
	if err != nil {
		return
	}
	converter = &enexToSN{
		snConverter: snConverter{},
	}
	if convResources, err = converter.convertToSN(origResources); err != nil {
		return
	}
	out = &SN{Items: convResources}
	err = writeResources(
		out,
		opts.OutputFilename,
		false,
		"standardnotes resources",
	)
	return
}

type toStandardNotes interface {
	convertToSN(in []entity.LinkID) (out []entity.LinkID, err error)
}

type (
	// snConverter is a base type for converting data from an outside service
	// into StandardNotes format.
	snConverter struct{}
	// edamToSN converts Evernote data into StandardNotes data using the EDAM
	// API.
	edamToSN struct{ snConverter }
	// enexToSN converts Evernote data into StandardNotes data using an ENEX
	// file.
	enexToSN struct{ snConverter }
)

func (c *snConverter) generateUUID() (string, error) {
	out, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c *edamToSN) convertToSN(in []entity.LinkID) ([]entity.LinkID, error) {
	notes := make([]entity.LinkID, 0)
	noteIDsByTagID := make(map[string][]string)
	noteIDsByNotebookID := make(map[string][]string)

	// process notes first so you can create references from notebooks, tags.
	for _, link := range in {
		item, ok := link.(*edam.Note)
		if !ok {
			continue
		}
		tagReferences := make([]sn.Reference, len(item.TagIDs))
		for j, tagID := range item.TagIDs {
			tagReferences[j] = sn.Reference{
				UUID:        tagID,
				ContentType: sn.ContentTypeTag,
			}
			if _, ok = noteIDsByTagID[tagID]; !ok {
				noteIDsByTagID[tagID] = make([]string, 0)
			}
			noteIDsByTagID[tagID] = append(noteIDsByTagID[tagID], item.ID)
		}
		if _, ok = noteIDsByNotebookID[item.NotebookID]; !ok {
			noteIDsByNotebookID[item.NotebookID] = make([]string, 0)
		}
		noteIDsByNotebookID[item.NotebookID] = append(
			noteIDsByNotebookID[item.NotebookID],
			item.ID,
		)
		text, xerr := extractNoteContent(item)
		if xerr != nil {
			return nil, xerr
		}
		notes = append(notes, &sn.Note{
			Item: sn.Item{
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
				ContentType: sn.ContentTypeNote,
				UUID:        item.ID,
				Content: struct {
					Title      string                 `json:"title"`
					References []sn.Reference         `json:"references"`
					Text       string                 `json:"text,omitempty"`
					AppData    map[string]interface{} `json:"appData,omitempty"`
				}{
					Title: item.Title,
					References: append(
						tagReferences,
						sn.Reference{
							UUID:        item.NotebookID,
							ContentType: sn.ContentTypeNotebook,
						},
					),
					Text: text,
					AppData: map[string]interface{}{
						"org.standardnotes.sn": &SNItemAppData{
							ClientUpdatedAt: &item.UpdatedAt,
						},
					},
				},
			},
		})
	}

	tags, notebooks := make([]entity.LinkID, 0), make([]entity.LinkID, 0)
	// after collecting note IDs, process notebooks, tags.
	for _, link := range in {
		switch item := link.(type) {
		case *edam.Note:
			// already processed
		case *edam.Tag:
			var noteReferences []sn.Reference
			if noteIDs, ok := noteIDsByTagID[item.ID]; !ok {
				noteReferences = make([]sn.Reference, 0)
			} else {
				noteReferences = make([]sn.Reference, len(noteIDs))
				for j, noteID := range noteIDs {
					noteReferences[j] = sn.Reference{
						UUID:        noteID,
						ContentType: sn.ContentTypeNote,
					}
				}
			}
			tags = append(tags, &sn.Tag{
				Item: sn.Item{
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
					ContentType: sn.ContentTypeTag,
					UUID:        item.ID,
					Content: struct {
						Title      string                 `json:"title"`
						References []sn.Reference         `json:"references"`
						Text       string                 `json:"text,omitempty"`
						AppData    map[string]interface{} `json:"appData,omitempty"`
					}{
						Title:      item.Name,
						References: noteReferences,
						AppData: map[string]interface{}{
							"evernote.com": &SNItemAppData{
								ParentID: item.ParentID,
							},
						},
					},
				},
			})
		case *edam.Notebook:
			var noteReferences []sn.Reference
			if noteIDs, ok := noteIDsByNotebookID[item.ID]; !ok {
				noteReferences = make([]sn.Reference, 0)
			} else {
				noteReferences = make([]sn.Reference, len(noteIDs))
				for j, noteID := range noteIDs {
					noteReferences[j] = sn.Reference{
						UUID:        noteID,
						ContentType: sn.ContentTypeNote,
					}
				}
			}
			notebooks = append(notebooks, &sn.Tag{
				Item: sn.Item{
					CreatedAt:   item.CreatedAt,
					UpdatedAt:   item.UpdatedAt,
					ContentType: sn.ContentTypeNotebook,
					UUID:        item.ID,
					Content: struct {
						Title      string                 `json:"title"`
						References []sn.Reference         `json:"references"`
						Text       string                 `json:"text,omitempty"`
						AppData    map[string]interface{} `json:"appData,omitempty"`
					}{
						Title:      item.Name,
						References: noteReferences,
						AppData: map[string]interface{}{
							"evernote.com": &SNItemAppData{
								OriginalContentType: "Notebook",
							},
						},
					},
				},
			})
		default:
			return nil, fmt.Errorf("%w; got %T", errTypeAssertion, item)
		}
	}

	out := make([]entity.LinkID, 0)
	out = append(out, notes...)
	out = append(out, tags...)
	out = append(out, notebooks...)
	return out, nil
}

func (c *enexToSN) convertToSN(in []entity.LinkID) (out []entity.LinkID, err error) {
	out = make([]entity.LinkID, len(in))
	tagsByName := make(map[string]*sn.Tag)
	// The official sntools implementation adds a list of tags at the end of the
	// output in the same order as read in the file. We can't rely on the
	// iteration order in a map, so a separate list is needed.
	listOfTags := make([]*sn.Tag, 0)

	for i, link := range in {
		enexNote, ok := link.(*enex.Note)
		if !ok {
			err = fmt.Errorf("%w; expected %T", errTypeAssertion, &enex.Note{})
			return
		}

		noteID, uerr := c.generateUUID()
		if uerr != nil {
			err = uerr
			return
		}
		tagReferences := make([]sn.Reference, len(enexNote.Tags))
		for j, tagName := range enexNote.Tags {
			if _, ok = tagsByName[tagName]; !ok {
				tagID, uerr := c.generateUUID()
				if uerr != nil {
					err = uerr
					return
				}
				tagsByName[tagName] = &sn.Tag{
					Item: sn.Item{
						CreatedAt:   time.Now().UTC(),
						UpdatedAt:   time.Now().UTC(),
						ContentType: sn.ContentTypeTag,
						UUID:        tagID,
						Content: struct {
							Title      string                 `json:"title"`
							References []sn.Reference         `json:"references"`
							Text       string                 `json:"text,omitempty"`
							AppData    map[string]interface{} `json:"appData,omitempty"`
						}{
							Title:      tagName,
							References: make([]sn.Reference, 0),
						},
					},
				}
				listOfTags = append(listOfTags, tagsByName[tagName])
			}
			tagReferences[j] = sn.Reference{
				UUID:        tagsByName[tagName].UUID,
				ContentType: sn.ContentTypeTag,
			}
			tagsByName[tagName].Content.References = append(
				tagsByName[tagName].Content.References,
				sn.Reference{
					UUID:        noteID,
					ContentType: sn.ContentTypeNote,
				},
			)
		}
		text, xerr := extractNoteContent(enexNote)
		if xerr != nil {
			err = xerr
			return
		}
		out[i] = &sn.Note{
			Item: sn.Item{
				CreatedAt:   enexNote.CreatedAt,
				UpdatedAt:   enexNote.UpdatedAt,
				ContentType: sn.ContentTypeNote,
				UUID:        noteID,
				Content: struct {
					Title      string                 `json:"title"`
					References []sn.Reference         `json:"references"`
					Text       string                 `json:"text,omitempty"`
					AppData    map[string]interface{} `json:"appData,omitempty"`
				}{
					Title:      enexNote.Title,
					References: tagReferences,
					Text:       text,
					AppData: map[string]interface{}{
						"org.standardnotes.sn": &SNItemAppData{
							ClientUpdatedAt: &enexNote.UpdatedAt,
						},
					},
				},
			},
		}
	}
	for i := range listOfTags {
		out = append(out, listOfTags[i])
	}
	return
}

var (
	enexLineBreakPattern = regexp.MustCompile(`/<br[^>]*>/g`)
	enexListItemPattern  = regexp.MustCompile(`/<li[^>]*>/g`)
)

type noteWithHTMLContent interface {
	HTMLContent() (out string, err error)
}

func extractNoteContent(note noteWithHTMLContent) (string, error) {
	content, err := note.HTMLContent()
	if err != nil {
		return "", err
	}
	out := enexLineBreakPattern.ReplaceAllString(content, "\n\n")
	out = enexListItemPattern.ReplaceAllString(out, "\n")
	return out, nil
}

// SNItemAppData is extra metadata attached to a StandardNotes Item that should
// be preserved between platforms or services.
type SNItemAppData struct {
	// ClientUpdatedAt is a pointer rather than value because of the omitempty
	// field tag. If it was a value, then omitempty would have no effect.
	ClientUpdatedAt *time.Time `json:"client_updated_at,omitempty"`
	// OriginalContentType is the kind of data in the origin service. For
	// example, StandardNotes does not have a Notebook type, the closest thing
	// is a Tag. This field is available to preserve this kind of metadata.
	OriginalContentType string `json:"original_content_type,omitempty"`
	// ParentID could be the ID of a parent resource in the original service.
	ParentID string `json:"parent_id,omitempty"`
}
