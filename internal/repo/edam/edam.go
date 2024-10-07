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
	"github.com/rafaelespinoza/notexfr/internal/entity"
	"github.com/rafaelespinoza/notexfr/internal/log"
	"github.com/rafaelespinoza/notexfr/internal/repo"
)

// EvernoteService enumerates different service environments when interacting
// with Evernote.
type EvernoteService uint8

// EvernoteServiceKey is the name of the context key that stores the EDAM
// credential configuration values.
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

// CredentialsConfig says where to find credentials and which to ones to use.
type CredentialsConfig struct {
	EnvFilename string
	ServiceEnv  EvernoteService
}

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

type store struct {
	edam.NoteStore
	token string
}

// initStore returns a pointer to the singleton store. When first called, it
// authenticates with the Evernote EDAM API using the credentials in the .env
// file. Upon success, it is initialized and cached for subsequent calls. An
// error may be returned while setting it up the first time.
func initStore(ctx context.Context) (*store, error) {
	if _TheStore != nil {
		return _TheStore, nil
	}
	var (
		err          error
		s            *store
		credsConf    CredentialsConfig
		credentials  userCredentials
		baseENClient *en.EvernoteClient
		userURLs     *edam.UserUrls
		userClient   *edam.UserStoreClient
		noteClient   *edam.NoteStoreClient
	)
	if val, ok := ctx.Value(EvernoteServiceKey).(CredentialsConfig); ok {
		credsConf = val
	} else {
		return nil, fmt.Errorf("could not read initial credentials config from context")
	}

	if credentials, err = loadEnv(credsConf); err != nil {
		return nil, err
	}
	log.Debug(ctx, map[string]any{"filename": credsConf.EnvFilename, "service_env": credsConf.ServiceEnv}, "env loaded")

	defer func() {
		log.Debug(ctx, map[string]any{
			"got_user_store": userClient != nil,
			"got_user_urls":  userURLs != nil,
			"got_note_store": noteClient != nil,
			"complete":       _TheStore != nil,
		}, "initStore status")
	}()

	baseENClient = en.NewClient(
		credentials.key,
		credentials.secret,
		en.EnvironmentType(credsConf.ServiceEnv),
	)
	if userClient, err = baseENClient.GetUserStore(); err != nil {
		return nil, makeError(err)
	}
	if userURLs, err = userClient.GetUserUrls(ctx, credentials.token); err != nil {
		return nil, makeError(err)
	}
	if noteClient, err = baseENClient.GetNoteStoreWithURL(userURLs.GetNoteStoreUrl()); err != nil {
		return nil, makeError(err)
	}
	s = &store{
		NoteStore: noteClient,
		token:     credentials.token,
	}
	_TheStore = s
	return _TheStore, nil
}

type userCredentials struct{ token, key, secret string }

// loadEnv produces user account credentials from environment variables. The
// credentials can be read from envfile or can be passed in from the command
// line. ie: `FOO=bar cmd args ...`. The imported library allows you to pass no
// arguments to read env vars from a file named .env. This might make sense for
// a web service because the working directory is fixed. I don't think this
// makes sense for a command because the working directory could be anywhere.
// Furthermore, this is sensitive information, that should be managed by the
// caller. Should the user choose to read from an env var file, force a
// non-empty name.
func loadEnv(credsConf CredentialsConfig) (userCredentials, error) {
	var out userCredentials
	if credsConf.EnvFilename != "" {
		if err := godotenv.Load(credsConf.EnvFilename); err != nil {
			return out, fmt.Errorf("could not load env vars; %v", err)
		}
	}
	if credsConf.ServiceEnv == EvernoteProductionService {
		out.token = os.Getenv(_ProductionTokenName)
		out.key = os.Getenv(_ProductionKeyName)
		out.secret = os.Getenv(_ProductionSecretName)
	} else {
		out.token = os.Getenv(_SandboxTokenName)
		out.key = os.Getenv(_SandboxKeyName)
		out.secret = os.Getenv(_SandboxSecretName)
	}
	return out, nil
}

// makeTimestamp converts an edam Timestamp to a regular golang time.Time in
// UTC. An edam.Timestamp should be in UTC, however it must be carefully
// converted since it's expressed in milliseconds since the Unix epoch.
func makeTimestamp(in edam.Timestamp) time.Time {
	return time.Unix(int64(in)/1000, 0).UTC()
}

func fmtTime(t time.Time) string { return t.Format(entity.Timeformat) }

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
