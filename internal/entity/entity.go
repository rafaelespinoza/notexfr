package entity

import "time"

// These are core entity types. They are modeled closely after types in actual
// Evernote, but tame the field naming and typing just a bit.
type (
	// A Notebook is a unique container for a set of notes and corresponds
	// to an Evernote Notebook.
	Notebook struct {
		// Name is the name of the Notebook itself.
		Name string
		// Stack is the name of the Notebook stack, if any.
		Stack string
		// CreatedAt is modeled after the edam.Notebook's ServiceCreated field.
		CreatedAt time.Time
		// UpdatedAt is modeled after the edam.Notebook's ServiceUpdated field.
		UpdatedAt time.Time

		// ID represents the GUID of the resource in Evernote.
		ID string
	}
	// A Note is a single note in the user's account and corresponds to an
	// Evernote Note.
	Note struct {
		// Title is the subject of a note.
		Title string
		// NotebookID associates the note to a notebook.
		NotebookID string
		// TagIDs is a list of tag IDs for the note.
		TagIDs []string
		// Tags is a list of tag names for the note.
		Tags []string
		// Content is the text content of the note.
		Content string
		// CreatedAt is modeled after the Created field.
		CreatedAt time.Time
		// UpdatedAt is modeled after the Updated field.
		UpdatedAt time.Time
		// Attributes is extra metadata about the note.
		Attributes *Attributes

		// ID represents the GUID of the resource in Evernote.
		ID string
	}
	// A Tag is a label to apply to a note and corresponds to an Evernote Tag.
	Tag struct {
		// Name is a unique user-defined name.
		Name string
		// ParentID is the GUID of the parent tag, if any.
		ParentID string

		// ID represents the GUID of the resource in Evernote.
		ID string
	}
)

// These are metadata entity types.
type (
	// Attributes is additional Note metadata and corresponds to Evernote
	// NoteAttributes.
	Attributes struct {
		ContentClass      string // TODO: maybe remove? curious to know what this is
		SourceApplication string // TODO: maybe remove? curious to know what this is
		Source            string
		SourceURL         string
	}
)

// A ServiceID identifies data as it's known in one service.
type ServiceID struct {
	Value string
}

// GetID returns Value.
func (i *ServiceID) GetID() string { return i.Value }

// SetID sets Value.
func (i *ServiceID) SetID(id string) { i.Value = id }
