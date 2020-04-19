// Package sn handles data to and from StandardNotes.
package sn

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rafaelespinoza/snbackfill/internal/entity"
)

// ReadConversionFile takes the output file of a conversion performed at:
// https://dashboard.standardnotes.org/tools and transforms the resources. The
// conversion file input is a flat array of items as JSON, where the content
// type is one of a few enumerable values, such as "Note", "Tag". This function
// groups items by content type into separate lists.
func ReadConversionFile(filename string) (notes, tags []entity.LinkID, err error) {
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
		case ContentTypeNote:
			item.Note.ServiceID = &entity.ServiceID{Value: item.Note.UUID}
			notes = append(notes, item.Note)
		case ContentTypeTag:
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
	contentType() ContentType
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

func (c *convfileItem) contentType() (out ContentType) {
	if c.Note != nil {
		out = ContentTypeNote
	} else if c.Tag != nil {
		out = ContentTypeTag
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
	var item Item
	if err = json.Unmarshal(data, &item); err != nil {
		return
	}
	switch item.ContentType {
	case ContentTypeNote:
		var note Note
		if err = json.Unmarshal(data, &note); err != nil {
			return
		}
		note.truncateTimes(time.Second)
		c.Note = &note
	case ContentTypeTag:
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

// An Item is an abstract type that should be embedded into a more concrete
// type such as a Note or a Tag. It contains metadata common to all
// StandardNotes item types and is loosely based off of
// https://docs.standardnotes.org/specification/sync/#items.
// There is a field to store text content, but is only relevant for notes.
type Item struct {
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ContentType ContentType `json:"content_type"`
	UUID        string      `json:"uuid"`
	Content     struct {
		Title      string                 `json:"title"`
		References []Reference            `json:"references"`
		Text       string                 `json:"text,omitempty"`
		AppData    map[string]interface{} `json:"appData,omitempty"`
	} `json:"content"`
	*entity.ServiceID `json:"-"`
}

func (i *Item) contentType() ContentType { return i.ContentType }

func (i *Item) truncateTimes(dur time.Duration) {
	i.CreatedAt = i.CreatedAt.Truncate(dur)
	i.UpdatedAt = i.UpdatedAt.Truncate(dur)
}

// Note is an item in a conversion file with a content type of Note.
type Note struct {
	Item
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
		if ref.ContentType == ContentTypeTag {
			currMembers[ref.UUID] = struct{}{}
		}
	}

	for _, id := range ids {
		if _, ok := currMembers[id]; ok {
			continue
		}
		n.Content.References = append(n.Content.References, Reference{
			UUID:        id,
			ContentType: ContentTypeTag,
		})
	}
	nextLen = len(n.Content.References)
	return
}

func fmtTime(t time.Time) string { return t.Format(entity.Timeformat) }

// Tag is an item in a conversion file with a content type of Tag.
type Tag struct {
	Item
}

// NewTag constructs a basic *Tag.
func NewTag(title string, created, updated time.Time) *Tag {
	item := Item{
		CreatedAt:   created,
		UpdatedAt:   updated,
		ContentType: ContentTypeTag,
		ServiceID:   &entity.ServiceID{Value: ""},
		Content: struct {
			Title      string                 `json:"title"`
			References []Reference            `json:"references"`
			Text       string                 `json:"text,omitempty"`
			AppData    map[string]interface{} `json:"appData,omitempty"`
		}{
			Title:      title,
			References: make([]Reference, 0),
			AppData:    make(map[string]interface{}),
		},
	}
	return &Tag{item}
}

func (t *Tag) LinkValues() []string { return []string{t.Content.Title} }

// A Reference is additional metadata for associating items.
type Reference struct {
	UUID        string      `json:"uuid"`
	ContentType ContentType `json:"content_type"`
}

var errContentTypeInvalid = errors.New("content_type invalid")

// ContentType describes the item.
type ContentType uint8

// These are the most common types of content, the first is a default value.
const (
	ContentTypeUnknown ContentType = iota
	ContentTypeNote
	ContentTypeTag
)

func (c ContentType) String() string {
	return [...]string{"", "Note", "Tag"}[c]
}

func (c *ContentType) MarshalJSON() (data []byte, err error) {
	data = []byte(`"` + c.String() + `"`)
	return
}

func (c *ContentType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	switch s {
	case "Note", "note":
		*c = ContentTypeNote
	case "Tag", "tag":
		*c = ContentTypeTag
	default:
		err = fmt.Errorf("%w; got %q", errContentTypeInvalid, s)
	}
	return
}
