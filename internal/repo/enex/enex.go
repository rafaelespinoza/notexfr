// Package enex specifically handles Evernote export files. The file is a
// specialized XML format called ENEX. More info can be found at
// https://evernote.com/blog/how-evernotes-xml-export-format-works/.
package enex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/macrat/go-enex"
	"github.com/rafaelespinoza/notexfr/internal/entity"
	"golang.org/x/net/html"
)

// File implements the local repository interface for enex files.
type File struct{}

// NewFileRepo constructs a File.
func NewFileRepo() (entity.RepoLocal, error) { return &File{}, nil }

const timeformat = "2006-01-02T15:04:05Z"

// ReadLocal reads and parses an enex file.
func (f *File) ReadLocal(ctx context.Context, r io.Reader) (out []entity.LinkID, err error) {
	var (
		parsed enex.EvernoteExportedXML
		note   entity.LinkID
	)

	if parsed, err = enex.ParseFromReader(r); err != nil {
		return
	}

	resources := make([]entity.LinkID, len(parsed.Notes))
	for i := range parsed.Notes {
		enexNote := parsed.Notes[i]
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
		n.CreatedAt.Format(entity.Timeformat),
		n.Title,
		n.UpdatedAt.Format(entity.Timeformat),
	}
}

// HTMLContent extracts the HTML from the note content.
func (n *Note) HTMLContent() (string, error) {
	root, err := html.Parse(strings.NewReader(n.Content))
	if err != nil {
		return "", err
	}
	// descend to <en-note> and capture its children.
	var curr *html.Node
	curr = root.LastChild
	if curr.Data != "html" {
		return "", fmt.Errorf("could not find node: html")
	}
	curr = curr.LastChild
	if curr.Data != "body" {
		return "", fmt.Errorf("could not find node: html.body")
	}
	curr = curr.FirstChild
	if curr.Data != "en-note" {
		return "", fmt.Errorf("could not find node: html.body.en-note")
	}
	var bld strings.Builder
	for curr = curr.FirstChild; curr != nil; curr = curr.NextSibling {
		if err = html.Render(&bld, curr); err != nil {
			return "", err
		}
	}
	return bld.String(), nil
}

func newNoteFromEnex(enexNote *enex.Note) (resource entity.LinkID, err error) {
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
			Content:   enexNote.Content.XML,
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
