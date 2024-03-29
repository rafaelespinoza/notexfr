package repo

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/rafaelespinoza/notexfr/internal/entity"
)

// Error should wrap the error from a third party library.
var Error = errors.New("repo error")

// FetchResources makes an API call for remote resources.
func FetchResources(ctx context.Context, repository entity.RepoRemote) (resources []entity.LinkID, err error) {
	resources, err = repository.FetchRemote(ctx)
	return
}

// ReadLocalFile takes a repository implementation, reads and parses results for
// that repository from a local file. To return tags pass in a Tags repo, for
// Notes pass in a Notes repo, etc.
func ReadLocalFile(ctx context.Context, repository entity.RepoLocal, filename string) ([]entity.LinkID, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filepath.Clean(filename)); err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return repository.ReadLocal(ctx, file)
}

// NewServiceID creates a ServiceID.
func NewServiceID(id string) entity.Resource { return &entity.ServiceID{Value: id} }
