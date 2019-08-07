package poststore

import (
	"time"

	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/types"
)

type Cursor struct {
	ID      string
	Created time.Time
}

type Store interface {
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

var EmptyCursor = Cursor{}
var ErrIDAlreadyExists = errors.New("A post with this ID already exists")
var ErrIDNotFound = errors.New("A post with this ID cannot be found")
