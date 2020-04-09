package edam

import (
	"context"
	"encoding/json"
	"io"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/rafaelespinoza/notexfr/internal/entity"
)

// Tags handles input/output for tags from the Evernote EDAM API.
type Tags struct{}

// NewTagsRepo constructs a Tags repository.
func NewTagsRepo() (entity.LocalRemoteRepo, error) { return &Tags{}, nil }

// FetchRemote gets Tags from the Evernote EDAM API.
func (n *Tags) FetchRemote(ctx context.Context) (out []entity.LinkID, err error) {
	var (
		s    *store
		tags []*edam.Tag
		i    int
		tag  *edam.Tag
		id   string
	)
	if s, err = initStore(ctx); err != nil {
		return
	}
	if tags, err = s.ListTags(ctx, s.token); err != nil {
		err = makeError(err)
		return
	}
	out = make([]entity.LinkID, len(tags))
	for i, tag = range tags {
		id = string(tag.GetGUID())
		out[i] = &Tag{
			Tag: &entity.Tag{
				Name:     tag.GetName(),
				ID:       id,
				ParentID: string(tag.GetParentGuid()),
			},
			ServiceID: &entity.ServiceID{Value: id},
		}
	}
	return
}

// ReadLocal reads and parses tags saved in a local JSON file.
func (n *Tags) ReadLocal(ctx context.Context, r io.Reader) (out []entity.LinkID, err error) {
	decoder := json.NewDecoder(r)
	var resources []*Tag
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

// Tag represents a tag in an Evernote EDAM API call. It also provides methods
// to match with a Tag in StandardNotes.
type Tag struct {
	*entity.Tag
	*entity.ServiceID
}

func (t *Tag) LinkValues() []string { return []string{t.Name} }
