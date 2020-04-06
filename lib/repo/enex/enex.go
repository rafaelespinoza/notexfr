// Package enex specifically handles Evernote export files.
package enex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rafaelespinoza/snbackfill/lib"
	"github.com/rafaelespinoza/snbackfill/lib/entity"

	"github.com/macrat/go-enex"
)

// File implements the local repository interface for enex files.
type File struct{}

// NewFileRepo constructs a File.
func NewFileRepo() (lib.RepoLocal, error) { return &File{}, nil }

const timeformat = "2006-01-02T15:04:05Z"

// ReadLocal reads and parses an enex file.
func (f *File) ReadLocal(ctx context.Context, r io.Reader) (out []lib.LinkID, err error) {
	var (
		parsed enex.EvernoteExportedXML
		note   lib.LinkID
	)

	if parsed, err = enex.ParseFromReader(r); err != nil {
		return
	}

	resources := make([]lib.LinkID, len(parsed.Notes))
	for i, enexNote := range parsed.Notes {
		if note, err = newNoteFromEnex(&enexNote); err != nil {
			return
		}
		resources[i] = note
	}
	out = resources
	return
}

// A Note is a note entity in an enex file.
type Note struct {
	*entity.Note
	*entity.ServiceID
}

func (n *Note) LinkValues() []string {
	return []string{
		n.CreatedAt.Format(lib.Timeformat),
		n.Title,
		n.UpdatedAt.Format(lib.Timeformat),
	}
}

func newNoteFromEnex(enexNote *enex.Note) (resource lib.LinkID, err error) {
	var createdAt, updatedAt time.Time
	if createdAt, err = time.Parse(timeformat, enexNote.CreatedAt.String()); err != nil {
		return
	}
	if updatedAt, err = time.Parse(timeformat, enexNote.UpdatedAt.String()); err != nil {
		return
	}
	var sourceURL string
	if enexNote.SourceURL != nil {
		sourceURL = enexNote.SourceURL.String()
	}
	resource = &Note{
		Note: &entity.Note{
			Title:     enexNote.Title,
			Tags:      enexNote.Tags,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Attributes: &entity.Attributes{
				Source:    enexNote.Source,
				SourceURL: sourceURL,
			},
		},
	}
	return
}

// FileOpts is named options for handling local enex files.
type FileOpts struct {
	Filename    string
	PrettyPrint bool
}

// ReadPrintFile provides a way to parse and inspect an Evernote export file (in
// ENEX format) as golang values.
func ReadPrintFile(ctx context.Context, opts *FileOpts) (err error) {
	var file *os.File
	var parsed enex.EvernoteExportedXML
	var exp *enex.EvernoteExportedXML

	if file, err = os.Open(opts.Filename); err != nil {
		return
	}
	if parsed, err = enex.ParseFromReader(file); err != nil {
		return
	}
	exp = &parsed

	if !opts.PrettyPrint {
		fmt.Println(exp)
		return
	}

	var data []byte
	if data, err = xml.MarshalIndent(exp, "", "    "); err != nil {
		return
	}
	fmt.Printf("%+v\n", string(data))

	return
}
