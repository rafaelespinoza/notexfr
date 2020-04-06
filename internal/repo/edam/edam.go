// Package edam handles interaction with the Evernote API, otherwise known as
// EDAM. Currently, a developer token is required to access an individual
// Evernote account. Read more at https://dev.evernote.com/doc/.
package edam

import (
	"context"
	"fmt"
	"os"
	"time"

	en "github.com/dreampuf/evernote-sdk-golang/client"
	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/joho/godotenv"
	lib "github.com/rafaelespinoza/snbackfill/internal"
	"github.com/rafaelespinoza/snbackfill/internal/repo"
)

// EvernoteService enumerates different service environments when interacting
// with Evernote.
type EvernoteService uint8

// EvernoteServiceKey is the name of the environment variable that stores the
// EvernoteService value.
const EvernoteServiceKey = "EVERNOTE_SERVICE"

const (
	// EvernoteSandboxService is a cautious default to operate on data in a
	// development environment, entirely separate from the production service.
	EvernoteSandboxService EvernoteService = iota
	// EvernoteProductionService signals that you want to use your production
	// Evernote account using production credentials.
	EvernoteProductionService
)

func (d EvernoteService) String() string {
	return [...]string{"SANDBOX", "PRODUCTION"}[d]
}

type store struct {
	noteClient *edamNoteStore

	credentials *evernoteCredentials
}

type evernoteCredentials struct{ Token, Key, Secret string }

const (
	_ProductionTokenName  = "EVERNOTE_PRODUCTION_TOKEN"
	_ProductionKeyName    = "EVERNOTE_PRODUCTION_KEY"
	_ProductionSecretName = "EVERNOTE_PRODUCTION_SECRET"
	_SandboxTokenName     = "EVERNOTE_SANDBOX_TOKEN"
	_SandboxKeyName       = "EVERNOTE_SANDBOX_KEY"
	_SandboxSecretName    = "EVERNOTE_SANDBOX_SECRET"
)

// MakeEnvFile initializes a template file to store evernote production and
// sandbox credentials for later usage as environment variables. If a file
// already exists, then it exits early with an error.
func MakeEnvFile(filename string) (err error) {
	_, err = os.Stat(filename)
	if err == nil {
		err = fmt.Errorf("env file already present at %q", filename)
		return
	} else if !os.IsNotExist(err) {
		return
	}
	err = godotenv.Write(
		map[string]string{
			_ProductionTokenName:  "",
			_ProductionKeyName:    "",
			_ProductionSecretName: "",
			_SandboxTokenName:     "",
			_SandboxKeyName:       "",
			_SandboxSecretName:    "",
		},
		filename,
	)
	return
}

var _TheStore *store

// initStore returns a pointer to the singleton store. When first called, it
// authenticates with the Evernote EDAM API using the credentials in the .env
// file. Upon success, it is initialized and cached for subsequent calls. An
// error may be returned while setting it up the first time.
func initStore(ctx context.Context) (*store, error) {
	if _TheStore != nil {
		return _TheStore, nil
	}
	var (
		err                error
		s                  *store
		evernoteServiceEnv EvernoteService
		credentials        evernoteCredentials
		baseENClient       *en.EvernoteClient
		userURLs           *edam.UserUrls
		userClient         *edam.UserStoreClient
		noteClient         *edam.NoteStoreClient
	)
	if err = godotenv.Load(); err != nil {
		return nil, err
	}
	if val, ok := ctx.Value(EvernoteServiceKey).(EvernoteService); ok {
		evernoteServiceEnv = val
	}
	if evernoteServiceEnv == EvernoteProductionService {
		credentials = evernoteCredentials{
			Token:  os.Getenv(_ProductionTokenName),
			Key:    os.Getenv(_ProductionKeyName),
			Secret: os.Getenv(_ProductionSecretName),
		}
	} else {
		credentials = evernoteCredentials{
			Token:  os.Getenv(_SandboxTokenName),
			Key:    os.Getenv(_SandboxKeyName),
			Secret: os.Getenv(_SandboxSecretName),
		}
	}
	s = &store{credentials: &credentials}
	baseENClient = en.NewClient(
		s.credentials.Key,
		s.credentials.Secret,
		en.EnvironmentType(evernoteServiceEnv),
	)
	if userClient, err = baseENClient.GetUserStore(); err != nil {
		return nil, err
	}
	if userURLs, err = userClient.GetUserUrls(ctx, s.credentials.Token); err != nil {
		return nil, err
	}
	if noteClient, err = baseENClient.GetNoteStoreWithURL(userURLs.GetNoteStoreUrl()); err != nil {
		return nil, err
	}
	s.noteClient = &edamNoteStore{
		NoteStore: noteClient,
		token:     s.credentials.Token,
	}
	_TheStore = s
	return _TheStore, nil
}

type edamNoteStore struct {
	edam.NoteStore
	token string
}

// makeTimestamp converts an edam Timestamp to a regular golang time.Time in
// UTC. An edam.Timestamp should be in UTC, however it must be carefully
// converted since it's expressed in milliseconds since the Unix epoch.
func makeTimestamp(in edam.Timestamp) time.Time {
	return time.Unix(int64(in)/1000, 0).UTC()
}

func fmtTime(t time.Time) string { return t.Format(lib.Timeformat) }

// makeError does some error wrapping for the EDAM API. See for details:
// - https://dev.evernote.com/doc/articles/error_handling.php
// - https://dev.evernote.com/doc/reference/Errors.html
func makeError(in error) (out error) {
	switch err := in.(type) {
	case *edam.EDAMUserException:
		out = fmt.Errorf(
			"client-side %w; type: %T, code: %q, parameter: %q",
			repo.Error, err, err.GetErrorCode().String(), err.GetParameter(),
		)
	case *edam.EDAMNotFoundException:
		out = fmt.Errorf(
			"client-side %w; type: %T, identifier: %q, key: %q",
			repo.Error, err, err.GetIdentifier(), err.GetKey(),
		)
	case *edam.EDAMSystemException:
		out = fmt.Errorf(
			"server-side %w; type: %T, code: %q, message: %q, rate limit duration: %d",
			repo.Error, err, err.GetErrorCode().String(), err.GetMessage(), err.GetRateLimitDuration(),
		)
	default:
		out = err
	}
	return
}
