// Package poststore provides storage and retrieval of posts.
package poststore

import (
	"time"

	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/types"
)

// Cursor is the struture used for pagination when listing posts.
//
// Posts when listing are ordered first by decreasing creation time (ie. most
// recent posts first), and then by ID.
type Cursor struct {
	ID      string
	Created time.Time
}

// Store is the common interface to all post stores.
//
// The store itself does not do any data validation (this is left to
// postservice), but simply stores and retrieves the data it's been given.
type Store interface {
	// Get retrieves a post by its ID. If the ID does not exist in the store,
	// ErrIDNotFound is returned.
	Get(id string) (types.Post, error)

	// Add adds a post to the store. If a post with this ID already exists, it
	// returns ErrIDAlreadyExists.
	Add(post types.Post) error

	// Update updates a given post in the store. If a post with the given ID
	// cannot be found, it returns ErrIDNotFound.
	Update(post types.Post) error

	// List lists the first n posts after the given cursor.
	//
	// EmptyCursor can be passed to list posts from the beginning.
	//
	// List returns a cursor that can be passed back to the next call for
	// continuing the iteration over the posts. When there are no more posts to
	// iterate, the returned cursor is EmptyCursor.
	List(c Cursor, n uint) (posts []types.Post, next Cursor, err error)
}

// EmptyCursor is the smallest cursor value.
//
// When used as a parameter to List, it tells that we want to retrieve the first
// page of results.
//
// When returned by List, it tells that there are no more results to list.
var EmptyCursor = Cursor{}

// ErrIDAlreadyExists is returned by Store.Create when trying to Add a post with
// an ID already present in the store.
var ErrIDAlreadyExists = errors.New("A post with this ID already exists")

// ErrIDNotFound is returned by Get or Update when trying to access an ID not
// present in the store.
var ErrIDNotFound = errors.New("A post with this ID cannot be found")
