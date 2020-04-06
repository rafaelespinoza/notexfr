package lib

import (
	"context"
	"io"
	"time"
)

// A Resource is a thin wrapper around a core entity type. An implementation
// should provide additional some additional methods to manage an unexported ID
// field so it can interact with a data store.
type Resource interface {
	// GetID should return an internal ID of the associated data store.
	GetID() string

	// SetID assigns the internal ID. It should only be used when a record is
	// first fetched from the data store.
	SetID(id string)
}

// A LocalRemoteRepo combines repository functionality for data in local files
// and data in a remote source.
type LocalRemoteRepo interface {
	RepoRemote
	RepoLocal
}

// A RepoRemote handles reading and parsing data, typically only available
// through an API call to a remote service.
type RepoRemote interface {
	FetchRemote(ctx context.Context) (out []LinkID, err error)
}

// A RepoLocal handles reading and parsing data that is typically on the local
// file system.
type RepoLocal interface {
	ReadLocal(ctx context.Context, reader io.Reader) (out []LinkID, err error)
}

// The ChainLink interface is for re-associating entities between services.
// Typically when the IDs have changed but some non-ID value has not, you can
// attempt to uniquely identify the same data as it is in multiple services by
// some non-primary key. For example, if you know two things in two separate
// services are Foos and you know each service has a unique constraint of the
// Bar property, this interface would return the Bar value in each Foo. If it's
// something that is likely to be unique in the common case, but does not
// preclude uniqueness, you can offer secondary, tertiary, n-ary values as the
// subsequent values. For example, your first key could be "created_at". If
// there are multiple possibilities, then compare further other field values
// such as "title", "updated_at" fields. This does not attempt to be a perfect
// solution, but is meant to provide an approximation for most cases.
type ChainLink interface {
	LinkValues() []string
}

// Timeformat is a standard layout for a time value.
const Timeformat = time.RFC3339

// LinkID is for attempting to identify the same data across different services.
type LinkID interface {
	Resource
	ChainLink
}
