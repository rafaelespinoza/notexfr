package interactor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rafaelespinoza/notexfr/internal/entity"
	"github.com/rafaelespinoza/notexfr/internal/log"
	"github.com/rafaelespinoza/notexfr/internal/repo"
	"github.com/rafaelespinoza/notexfr/internal/repo/edam"
	"github.com/rafaelespinoza/notexfr/internal/repo/enex"
)

// FetchWriteParams is a set of named arguments for fetching remote resources
// and/or writing results to a local file.
type FetchWriteParams struct {
	InputFilename    string
	OutputFilename   string
	Timeout          time.Duration
	NotesQueryParams *edam.NotesRemoteQueryParams
}

// FetchWriteNotebooks gets Notebooks from your Evernote account and writes the
// results to a local JSON file.
func FetchWriteNotebooks(ctx context.Context, opts *FetchWriteParams) (err error) {
	var repository entity.LocalRemoteRepo
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	if repository, err = edam.NewNotebooksRepo(); err != nil {
		return
	}
	err = fetchWriteResource(ctx, repository, opts, "Notebooks")
	return
}

// FetchWriteTags gets Tags from your Evernote account and writes the results
// to a local JSON file.
func FetchWriteTags(ctx context.Context, opts *FetchWriteParams) (err error) {
	var repository entity.LocalRemoteRepo
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	if repository, err = edam.NewTagsRepo(); err != nil {
		return
	}
	err = fetchWriteResource(ctx, repository, opts, "Tags")
	return
}

// FetchWriteNotes gets Notes from your Evernote account and writes the results
// to a local JSON file.
func FetchWriteNotes(ctx context.Context, opts *FetchWriteParams) (err error) {
	var repository entity.LocalRemoteRepo
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	if repository, err = edam.NewNotesRepo(opts.NotesQueryParams); err != nil {
		return
	}
	err = fetchWriteResource(ctx, repository, opts, "Notes")
	return
}

// WriteENEXToJSON converts an Evernote export file to JSON.
func WriteENEXToJSON(ctx context.Context, opts *FetchWriteParams) (err error) {
	var repository entity.RepoLocal
	var resources []entity.LinkID
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	if repository, err = enex.NewFileRepo(); err != nil {
		return
	}
	if resources, err = readLocalFile(ctx, repository, opts.InputFilename); err != nil {
		return
	}
	err = writeResources(resources, opts.OutputFilename, "ENEX export items")
	return
}

func fetchWriteResource(ctx context.Context, repository entity.LocalRemoteRepo, opts *FetchWriteParams, name string) (err error) {
	var resources []entity.LinkID
	if resources, err = fetchResources(ctx, repository, name); err != nil {
		return
	}
	err = writeResources(resources, opts.OutputFilename, name)
	return
}

func fetchResources(ctx context.Context, repository entity.LocalRemoteRepo, name string) (resources []entity.LinkID, err error) {
	if resources, err = repo.FetchResources(ctx, repository); err != nil {
		return
	}
	log.Info(ctx, map[string]any{"count": len(resources)}, "fetched "+name)
	return
}

// writeResources marshalizes resources to JSON and writes to a local file. If
// filename is empty, then it prints to standard output.
func writeResources(resources interface{}, filename string, name string) (err error) {
	data, err := json.Marshal(resources)
	if err != nil {
		return
	}
	if filename == "" {
		fmt.Println(string(data))
		return
	}
	err = os.WriteFile(filename, data, os.FileMode(0644))
	if err != nil {
		return
	}
	log.Info(context.TODO(), map[string]any{"filename": filename, "resource_type": name}, "wrote JSON data to file")
	return
}

func readLocalFile(ctx context.Context, repository entity.RepoLocal, filename string) ([]entity.LinkID, error) {
	return repo.ReadLocalFile(ctx, repository, filename)
}

// MakeEDAMEnvFile creates a file at filename with environment variables for
// the Evernote API unless it already exists.
func MakeEDAMEnvFile(envfile string) (err error) {
	if envfile == "" {
		err = fmt.Errorf("envfile is required")
		return
	}
	if err = edam.MakeEnvFile(envfile); err != nil {
		return
	}

	log.Info(context.TODO(), map[string]any{"filename": envfile}, "wrote envfile")
	return
}
