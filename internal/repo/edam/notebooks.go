package edam

import (
	"context"
	"encoding/json"
	"io"

	"github.com/rafaelespinoza/notexfr/internal/entity"
)

// Notebooks handles input/output for notebooks from the Evernote EDAM API.
type Notebooks struct{}

// NewNotebooksRepo constructs a Notebooks repository.
func NewNotebooksRepo() (entity.LocalRemoteRepo, error) { return &Notebooks{}, nil }

// FetchRemote gets Notebooks from the Evernote EDAM API.
func (n *Notebooks) FetchRemote(ctx context.Context) (out []entity.LinkID, err error) {
	var s *store
	var id string
	if s, err = initStore(ctx); err != nil {
		return
	}
	notebooks, err := s.noteClient.ListNotebooks(ctx, s.noteClient.token)
	if err != nil {
		err = makeError(err)
		return
	}
	out = make([]entity.LinkID, len(notebooks))
	for i, notebook := range notebooks {
		id = string(notebook.GetGUID())
		out[i] = &Notebook{
			Notebook: &entity.Notebook{
				ID:        id,
				Name:      notebook.GetName(),
				Stack:     notebook.GetStack(),
				CreatedAt: makeTimestamp(notebook.GetServiceCreated()),
				UpdatedAt: makeTimestamp(notebook.GetServiceUpdated()),
			},
			ServiceID: &entity.ServiceID{Value: id},
		}
	}
	return
}

// ReadLocal reads and parses notebooks saved in a local JSON file.
func (n *Notebooks) ReadLocal(ctx context.Context, r io.Reader) (out []entity.LinkID, err error) {
	decoder := json.NewDecoder(r)
	var resources []*Notebook
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

// Notebook represents a notebook in an Evernote EDAM API call. Though it
// provides methods to match with a resource in StandardNotes, there is no
// equivalent to Notebook in StandardNotes; as of this writing, the closest
// thing is a Tag.
type Notebook struct {
	*entity.Notebook
	*entity.ServiceID
}

// NewNotebook constructs a basic *Notebook. It doesn't set all possible fields.
func NewNotebook(name string) *Notebook {
	return &Notebook{
		Notebook:  &entity.Notebook{Name: name},
		ServiceID: &entity.ServiceID{Value: ""},
	}
}

func (n *Notebook) LinkValues() []string { return []string{n.Name} }
