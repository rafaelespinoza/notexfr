package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/entity"
)

// Error should wrap the error from a third party library.
var Error = errors.New("repo error")

// FetchResources makes an API call for remote resources.
func FetchResources(ctx context.Context, repository lib.RepoRemote) (resources []lib.LinkID, err error) {
	resources, err = repository.FetchRemote(ctx)
	return
}

// WriteResourcesJSON marshalizes resources to JSON and writes to a local file.
// If filename is empty, then it prints to standard output.
func WriteResourcesJSON(resources []lib.LinkID, filename string) (err error) {
	data, err := json.Marshal(resources)
	if err != nil {
		return
	}
	if filename == "" {
		fmt.Println(string(data))
		return
	}
	err = ioutil.WriteFile(filename, data, os.FileMode(0644))
	return
}

// ReadLocalFile takes a repository implementation, reads and parses results for
// that repository from a local file. To return tags pass in a Tags repo, for
// Notes pass in a Notes repo, etc.
func ReadLocalFile(ctx context.Context, repository lib.RepoLocal, filename string) ([]lib.LinkID, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		return nil, err
	}
	defer file.Close()
	return repository.ReadLocal(ctx, file)
}

// NewServiceID returns an *entity.ServiceID so use cases don't need to directly
// depend on the entity package.
func NewServiceID(id string) lib.Resource { return &entity.ServiceID{Value: id} }
