package postservice

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/abustany/back-message-board/pkg/poststore"
	"github.com/abustany/back-message-board/pkg/types"
)

type Service interface {
	// Get gets a post from the store. If the ID does not exist in the store, an
	// error is returned.
	Get(id string) (types.Post, error)

	// Add adds a new post to the store.
	Add(post types.Post) error

	// Update updates an existing post (identified by its ID) in the store.
	//
	// Partial updates are supported by only setting the fields that should be
	// updated in the post.
	Update(post types.Post) error

	// List returns the n most recent posts in the store, starting at the given
	// cursor. If the cursor is an empty string, the first page is returned.
	// n can be set to 0 to get the default page size.
	List(cursor string, n uint) (posts []types.Post, nextCursor string, err error)
}

type postService struct {
	store poststore.Store
}

const DefaultPageSize = 100

const MaxAuthorLength = 256
const MaxEmailLength = 256
const MaxMessageLength = 2048
const MaxPageSize = 100

var ErrInvalidID = &userError{errors.New("Invalid ID (should not be empty)")}
var ErrInvalidAuthor = &userError{errors.Errorf("Invalid author (should not be empty or longer than %d characters)", MaxAuthorLength)}
var ErrInvalidEmail = &userError{errors.Errorf("Invalid email (should not be empty or longer than %d characters)", MaxEmailLength)}
var ErrInvalidMessage = &userError{errors.Errorf("Invalid message (should not be longer than %d characters)", MaxMessageLength)}
var ErrInvalidCursor = &userError{errors.New("Invalid cursor")}
var ErrInvalidPageSize = &userError{errors.Errorf("Invalid page size (should not be larger than %d)", MaxPageSize)}

func New(store poststore.Store) Service {
	return &postService{store}
}

func validatePost(post types.Post, newPost bool) error {
	if !newPost && post.ID == "" {
		return ErrInvalidID
	}

	if (newPost && post.Author == "") || len(post.Author) > MaxAuthorLength {
		return ErrInvalidAuthor
	}

	// FIXME: not really a proper email address validation...
	if (newPost && post.Email == "") || len(post.Email) > MaxEmailLength {
		return ErrInvalidEmail
	}

	if len(post.Message) > MaxMessageLength {
		return ErrInvalidMessage
	}

	return nil
}

func (s *postService) Get(id string) (types.Post, error) {
	post, err := s.store.Get(id)

	if err == poststore.ErrIDNotFound {
		err = &userError{err}
	}

	return post, err
}

func (s *postService) Add(post types.Post) error {
	if err := validatePost(post, true); err != nil {
		return errors.Wrap(err, "Invalid post data")
	}

	post.ID = uuid.NewV4().String()
	post.Created = time.Now()

	return errors.Wrap(s.store.Add(post), "Error while adding post to store")
}

func (s *postService) Update(post types.Post) error {
	if err := validatePost(post, false); err != nil {
		return errors.Wrap(err, "Invalid post data")
	}

	err := s.store.Update(post)

	if err == poststore.ErrIDNotFound {
		err = &userError{err}
	}

	return errors.Wrap(err, "Error while updating post in store")
}

func encodeCursor(cursor poststore.Cursor) (string, error) {
	if cursor == poststore.EmptyCursor {
		return "", nil
	}

	jsonEncoded, err := json.Marshal(cursor)

	if err != nil {
		return "", errors.Wrap(err, "Error while encoding cursor to JSON")
	}

	return base64.URLEncoding.EncodeToString(jsonEncoded), nil
}

func decodeCursor(cursor string) (poststore.Cursor, error) {
	if cursor == "" {
		return poststore.EmptyCursor, nil
	}

	jsonEncoded, err := base64.URLEncoding.DecodeString(cursor)

	if err != nil {
		return poststore.Cursor{}, errors.Wrap(err, "Error while decoding base64")
	}

	var decoded poststore.Cursor

	if err := json.Unmarshal(jsonEncoded, &decoded); err != nil {
		return poststore.Cursor{}, errors.Wrap(err, "Error while decoding JSON")
	}

	if _, err := uuid.FromString(decoded.ID); err != nil {
		return poststore.Cursor{}, errors.Wrap(err, "Error while decoding cursor ID")
	}

	return decoded, nil
}

func (s *postService) List(cursor string, n uint) ([]types.Post, string, error) {
	if n > MaxPageSize {
		return nil, "", ErrInvalidPageSize
	}

	if n == 0 {
		n = DefaultPageSize
	}

	decodedCursor, err := decodeCursor(cursor)

	if err != nil {
		return nil, "", ErrInvalidCursor
	}

	posts, nextCursor, err := s.store.List(decodedCursor, n)

	if err != nil {
		return nil, "", errors.Wrap(err, "Error while listing posts")
	}

	nextCursorStr, err := encodeCursor(nextCursor)

	if err != nil {
		return nil, "", errors.Wrap(err, "Error while encoding next cursor")
	}

	return posts, nextCursorStr, nil
}
