// Package sn handles data to and from StandardNotes.
package sn

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/entity"
)

// ReadConversionFile takes the output file of a conversion performed at:
// https://dashboard.standardnotes.org/tools and transforms the resources. The
// conversion file input is a flat array of items as JSON, where the content
// type is one of a few enumerable values, such as "Note", "Tag". This function
// groups items by content type into separate lists.
func ReadConversionFile(filename string) (notes, tags []lib.LinkID, err error) {
	var (
		file      *os.File
		decoder   *json.Decoder
		metadatas struct{ Items []convfileItem }
	)
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer file.Close()
	decoder = json.NewDecoder(file)
	if err = decoder.Decode(&metadatas); err != nil {
		return
	}

	for _, item := range metadatas.Items {
		switch typ := item.contentType(); typ {
		case contentTypeNote:
			item.Note.ServiceID = &entity.ServiceID{Value: item.Note.UUID}
			notes = append(notes, item.Note)
		case contentTypeTag:
			item.Tag.ServiceID = &entity.ServiceID{Value: item.Tag.UUID}
			tags = append(tags, item.Tag)
		default:
			err = fmt.Errorf("%w; got: %q", errContentTypeInvalid, typ)
			return
		}
	}
	return
}

type contentTypeOption interface {
	contentType() contentType
}

// A convfileItem helps parse an item in an input file, which contains an items
// field. Depending on the content_type, an item is either a Note or a Tag.
type convfileItem struct {
	contentTypeOption
	*Note
	*Tag
}

var (
	_ json.Unmarshaler  = (*convfileItem)(nil)
	_ json.Marshaler    = (*convfileItem)(nil)
	_ contentTypeOption = (*convfileItem)(nil)
)

func (c *convfileItem) contentType() (out contentType) {
	if c.Note != nil {
		out = contentTypeNote
	} else if c.Tag != nil {
		out = contentTypeTag
	}
	return
}

func (c *convfileItem) MarshalJSON() (data []byte, err error) {
	if c.Note != nil {
		data, err = json.Marshal(c.Note)
	} else if c.Tag != nil {
		data, err = json.Marshal(c.Tag)
	} else {
		err = errContentTypeInvalid
	}
	return
}

func (c *convfileItem) UnmarshalJSON(data []byte) (err error) {
	var metadata Metadata
	if err = json.Unmarshal(data, &metadata); err != nil {
		return
	}
	switch metadata.ContentType {
	case contentTypeNote:
		var note Note
		if err = json.Unmarshal(data, &note); err != nil {
			return
		}
		note.truncateTimes(time.Second)
		c.Note = &note
	case contentTypeTag:
		var tag Tag
		if err = json.Unmarshal(data, &tag); err != nil {
			return
		}
		tag.truncateTimes(time.Second)
		c.Tag = &tag
	default:
		err = errContentTypeInvalid
	}
	return
}

// Metadata is relevant data common to all StandardNote item types. It's
// loosely based off https://docs.standardnotes.org/specification/sync/#items.
type Metadata struct {
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	ContentType       contentType `json:"content_type"`
	UUID              string      `json:"uuid"`
	*entity.ServiceID `json:"-"`
}

func (m *Metadata) contentType() contentType { return m.ContentType }

func (m *Metadata) truncateTimes(dur time.Duration) {
	m.CreatedAt = m.CreatedAt.Truncate(dur)
	m.UpdatedAt = m.UpdatedAt.Truncate(dur)
}

// Note is an item in a conversion file with a content type of Note.
type Note struct {
	Metadata
	Content struct {
		Title      string `json:"title"`
		References []references
	} `json:"content"`
}

func (n *Note) LinkValues() []string {
	return []string{
		fmtTime(n.CreatedAt),
		n.Content.Title,
		fmtTime(n.UpdatedAt),
	}
}

func (n *Note) AppendTags(ids ...string) (nextLen int) {
	currMembers := make(map[string]struct{})
	for _, ref := range n.Content.References {
		if ref.ContentType == contentTypeTag {
			currMembers[ref.UUID] = struct{}{}
		}
	}

	for _, id := range ids {
		if _, ok := currMembers[id]; ok {
			continue
		}
		n.Content.References = append(n.Content.References, references{
			UUID:        id,
			ContentType: contentTypeTag,
		})
	}
	nextLen = len(n.Content.References)
	return
}

func fmtTime(t time.Time) string { return t.Format(lib.Timeformat) }

// Tag is an item in a conversion file with a content type of Tag.
type Tag struct {
	Metadata
	Content struct {
		Title      string `json:"title"`
		References []references
	} `json:"content"`
}

// NewTag constructs a basic *ConvfileTag. It doesn't set all possible fields.
func NewTag(title string, created, updated time.Time) *Tag {
	return &Tag{
		Metadata: Metadata{
			CreatedAt:   created,
			UpdatedAt:   updated,
			ContentType: contentTypeTag,
			ServiceID:   &entity.ServiceID{Value: ""},
		},
		Content: struct {
			Title      string `json:"title"`
			References []references
		}{
			Title: title,
		},
	}
}

func (t *Tag) LinkValues() []string { return []string{t.Content.Title} }

type references struct {
	UUID        string      `json:"uuid"`
	ContentType contentType `json:"content_type"`
}

var errContentTypeInvalid = errors.New("content_type invalid")

type contentType uint8

const (
	contentTypeUnknown contentType = iota
	contentTypeNote
	contentTypeTag
)

func (c contentType) String() string {
	return [...]string{"", "Note", "Tag"}[c]
}

func (c *contentType) MarshalJSON() (data []byte, err error) {
	data = []byte(`"` + c.String() + `"`)
	return
}

func (c *contentType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	switch s {
	case "Note", "note":
		*c = contentTypeNote
	case "Tag", "tag":
		*c = contentTypeTag
	default:
		err = fmt.Errorf("%w; got %q", errContentTypeInvalid, s)
	}
	return
}
