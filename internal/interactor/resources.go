package interactor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/rafaelespinoza/notexfr/internal/entity"
	"github.com/rafaelespinoza/notexfr/internal/repo"
	"github.com/rafaelespinoza/notexfr/internal/repo/edam"
	"github.com/rafaelespinoza/notexfr/internal/repo/enex"
)

// FetchWriteOptions is a set of named arguments for fetching remote resources
// and/or writing results to a local file.
type FetchWriteOptions struct {
	InputFilename    string
	OutputFilename   string
	Timeout          time.Duration
	Verbose          bool
	NotesQueryParams *edam.NotesRemoteQueryParams
}

// FetchWriteNotebooks gets Notebooks from your Evernote account and writes the
// results to a local JSON file.
func FetchWriteNotebooks(ctx context.Context, opts *FetchWriteOptions) (err error) {
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
func FetchWriteTags(ctx context.Context, opts *FetchWriteOptions) (err error) {
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
func FetchWriteNotes(ctx context.Context, opts *FetchWriteOptions) (err error) {
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
func WriteENEXToJSON(ctx context.Context, opts *FetchWriteOptions) (err error) {
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
	err = writeResources(resources, opts.OutputFilename, opts.Verbose, "ENEX export items")
	return
}

func fetchWriteResource(ctx context.Context, repository entity.LocalRemoteRepo, opts *FetchWriteOptions, name string) (err error) {
	var resources []entity.LinkID
	if resources, err = fetchResources(ctx, repository, opts, name); err != nil {
		return
	}
	err = writeResources(resources, opts.OutputFilename, opts.Verbose, name)
	return
}

func fetchResources(ctx context.Context, repository entity.LocalRemoteRepo, opts *FetchWriteOptions, name string) (resources []entity.LinkID, err error) {
	if resources, err = repo.FetchResources(ctx, repository); err != nil {
		return
	}
	if opts.Verbose {
		fmt.Printf("fetched %d %s\n", len(resources), name)
	}
	return
}

// writeResources marshalizes resources to JSON and writes to a local file. If
// filename is empty, then it prints to standard output.
func writeResources(resources interface{}, filename string, verbose bool, name string) (err error) {
	data, err := json.Marshal(resources)
	if err != nil {
		return
	}
	if filename == "" {
		fmt.Println(string(data))
		return
	}
	err = ioutil.WriteFile(filename, data, os.FileMode(0644))
	if err != nil {
		return
	}
	if verbose {
		fmt.Printf("wrote %s to %q\n", name, filename)
	}
	return
}

func readLocalFile(ctx context.Context, repository entity.RepoLocal, filename string) ([]entity.LinkID, error) {
	return repo.ReadLocalFile(ctx, repository, filename)
}

// MakeEDAMEnvFile creates a file at filename with environment variables for
// the Evernote API unless it already exists.
func MakeEDAMEnvFile(filename string) (err error) {
	if err = edam.MakeEnvFile(filename); err != nil {
		return
	}
	fmt.Printf("ok, wrote to %q\n", filename)
	return
}
