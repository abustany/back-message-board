package postservice_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/postservice"
	"github.com/abustany/back-message-board/pkg/poststore"
	"github.com/abustany/back-message-board/pkg/types"
)

func TestPostService(t *testing.T) {
	withService := func(f func(*testing.T, postservice.Service)) func(*testing.T) {
		return func(t *testing.T) {
			store, err := poststore.NewMemoryPostStore()

			if err != nil {
				t.Fatalf("Error while creating post store: %s", err)
			}

			f(t, postservice.New(store))
		}
	}

	t.Run("Add (validation)", withService(testAddInvalid))
	t.Run("Add", withService(testAdd))

	t.Run("Update (validation)", withService(testUpdateInvalid))
	t.Run("Update", withService(testUpdate))

	t.Run("List (validation)", withService(testListValidation))
	t.Run("List", withService(testList))
}

func repeatStringUntil(s string, sizeAtLeast int) string {
	buffer := bytes.Buffer{}

	for buffer.Len() < sizeAtLeast {
		buffer.WriteString(s)
	}

	return buffer.String()
}

func expectError(t *testing.T, e, expected error) {
	for {
		cause := errors.Cause(e)

		if e == cause {
			break
		}

		e = cause
	}

	if e != expected {
		t.Errorf("Expected error %v, got %v", expected, e)
	}
}

const validAuthor = "John"

var tooLongAuthor = repeatStringUntil(validAuthor, postservice.MaxAuthorLength+1)

const validEmail = "john@domain.com"

var tooLongEmail = repeatStringUntil(validEmail, postservice.MaxEmailLength+1)

const validMessage = "Hello world!"

var tooLongMessage = repeatStringUntil(validMessage, postservice.MaxMessageLength+1)

func testValidation(t *testing.T, service postservice.Service, testEmpty bool, f func(*testing.T, types.Post) error) {
	if testEmpty {
		t.Run("Empty author", func(t *testing.T) {
			expectError(t, f(t, types.Post{
				Author:  "",
				Email:   validEmail,
				Message: validMessage,
			}), postservice.ErrInvalidAuthor)
		})

		t.Run("Empty email", func(t *testing.T) {
			expectError(t, f(t, types.Post{
				Author:  validAuthor,
				Email:   "",
				Message: validMessage,
			}), postservice.ErrInvalidEmail)
		})
	}

	t.Run("Too long author", func(t *testing.T) {
		expectError(t, f(t, types.Post{
			Author:  tooLongAuthor,
			Email:   validEmail,
			Message: validMessage,
		}), postservice.ErrInvalidAuthor)
	})

	t.Run("Too long email", func(t *testing.T) {
		expectError(t, f(t, types.Post{
			Author:  validAuthor,
			Email:   tooLongEmail,
			Message: validMessage,
		}), postservice.ErrInvalidEmail)
	})

	t.Run("Too long message", func(t *testing.T) {
		expectError(t, f(t, types.Post{
			Author:  validAuthor,
			Email:   validEmail,
			Message: tooLongMessage,
		}), postservice.ErrInvalidMessage)
	})

	// Check that no posts were actually added to the store
	posts, _, err := service.List("", 100)

	if err != nil {
		t.Fatalf("List returned an error: %s", err)
	}

	if len(posts) != 0 {
		t.Errorf("No posts should have been added to the store")
	}
}

func testAddInvalid(t *testing.T, service postservice.Service) {
	testValidation(t, service, true, func(t *testing.T, post types.Post) error {
		return service.Add(post)
	})
}

func listPosts(t *testing.T, service postservice.Service, expectedNumber int) []types.Post {
	posts, cursor, err := service.List("", 100)

	if err != nil {
		t.Fatalf("List returned an error: %s", err)
	}

	if cursor != "" {
		t.Fatalf("List didn't return an empty cursor")
	}

	if len(posts) != expectedNumber {
		t.Fatalf("Unexpected number of posts, got %d, expected %d", len(posts), expectedNumber)
		return nil
	}

	return posts
}

func testAdd(t *testing.T, service postservice.Service) {
	post := types.Post{
		ID:      "should be ignored",
		Author:  validAuthor,
		Email:   validEmail,
		Created: time.Now().Add(time.Hour), // should be ignored too
		Message: validMessage,
	}

	if err := service.Add(post); err != nil {
		t.Errorf("Add returned an error: %s", err)
	}

	posts := listPosts(t, service, 1)
	saved := posts[0]

	if saved.ID == post.ID {
		t.Errorf("Provided ID should have been ignored")
	}

	if saved.Created == post.Created {
		t.Errorf("Provided Created should have been ignored")
	}

	if saved.Author != post.Author {
		t.Errorf("Unexpected author, got %s, expected %s", saved.Author, post.Author)
	}

	if saved.Email != post.Email {
		t.Errorf("Unexpected email, got %s, expected %s", saved.Email, post.Email)
	}

	if saved.Message != post.Message {
		t.Errorf("Unexpected message, got %s, expected %s", saved.Message, post.Message)
	}
}

func testUpdateInvalid(t *testing.T, service postservice.Service) {
	testValidation(t, service, false, func(t *testing.T, post types.Post) error {
		post.ID = "ID"
		return service.Update(post)
	})

	post := types.Post{
		ID: "does not exist",
	}

	if err := service.Update(post); err == nil {
		t.Errorf("Expected an error when updating a non existing post")
	} else if !postservice.IsUserError(err) {
		t.Errorf("Updating a non existing post should be a user error")
	}
}

func testUpdate(t *testing.T, service postservice.Service) {
	post := types.Post{
		Author:  validAuthor,
		Email:   validEmail,
		Message: validMessage,
	}

	if err := service.Add(post); err != nil {
		t.Fatalf("Add returned an error: %s", err)
	}

	posts := listPosts(t, service, 1)
	post.ID = posts[0].ID
	post.Created = posts[0].Created

	t.Run("Update all fields", func(t *testing.T) {
		post.Author = post.Author + "x"
		post.Email = post.Email + "x"
		post.Created = post.Created.Add(time.Hour)
		post.Message = post.Message + "x"

		if err := service.Update(post); err != nil {
			t.Errorf("Update returned an error: %s", err)
		}

		posts := listPosts(t, service, 1)

		if !posts[0].Equal(post) {
			t.Errorf("Update didn't update: got %+v, expected %+v", posts[0], post)
		}
	})

	t.Run("Partial update", func(t *testing.T) {
		post.Author = post.Author + "y"

		patch := types.Post{
			ID:     post.ID,
			Author: post.Author,
		}

		if err := service.Update(patch); err != nil {
			t.Errorf("Update returned an error: %s", err)
		}

		posts := listPosts(t, service, 1)

		if !posts[0].Equal(post) {
			t.Errorf("Partial update didn't update: got %+v, expected %+v", posts[0], post)
		}
	})
}

func testListValidation(t *testing.T, service postservice.Service) {
	validationCheck := func(cursor string, pageSize uint, expectedError error) {
		posts, next, err := service.List(cursor, pageSize)

		if err == nil {
			t.Errorf("List didn't return an error")
		} else if err != expectedError {
			t.Errorf("Unexpected error: got %v, expected, %v", err, postservice.ErrInvalidCursor)
		}

		if !postservice.IsUserError(expectedError) {
			t.Errorf("Expected error %v is not a user error", expectedError)
		}

		if len(posts) != 0 {
			t.Errorf("List should have returned no posts")
		}

		if next != "" {
			t.Errorf("List should have returned an empty cursor")
		}
	}

	t.Run("Invalid cursor", func(t *testing.T) {
		validationCheck("not a valid cursor", postservice.MaxPageSize, postservice.ErrInvalidCursor)
	})

	t.Run("Too big page size", func(t *testing.T) {
		validationCheck("", postservice.MaxPageSize+1, postservice.ErrInvalidPageSize)
	})
}

func makePost(idx uint) types.Post {
	idxStr := fmt.Sprintf("%04d", idx)

	return types.Post{
		Author:  validAuthor + idxStr,
		Email:   validEmail + idxStr,
		Message: validMessage + idxStr,
	}
}

func listExpect(t *testing.T, service postservice.Service, cursor string, pageSize uint, expectedPosts []types.Post, expectEmptyCursor bool) string {
	posts, next, err := service.List(cursor, pageSize)

	if err != nil {
		t.Errorf("List returned an error: %s", err)
	}

	if len(posts) != len(expectedPosts) {
		t.Errorf("List returned %d posts, expected %d", len(posts), len(expectedPosts))
		return next
	}

	for i := range expectedPosts {
		posts[i].ID = ""
		posts[i].Created = time.Time{}

		if !posts[i].Equal(expectedPosts[i]) {
			t.Errorf("List returned an unexpected post at index %d: got %+v, expected %+v", i, posts[i], expectedPosts[i])
		}
	}

	return next
}

func testList(t *testing.T, service postservice.Service) {
	t.Run("Empty store", func(t *testing.T) {
		posts, next, err := service.List("", postservice.MaxPageSize)

		if err != nil {
			t.Errorf("List returned an error: %s", err)
		}

		if next != "" {
			t.Error("List didn't return an empty cursor")
		}

		if len(posts) != 0 {
			t.Error("List returned some posts")
		}
	})

	const nPosts = 100

	for i := uint(0); i < nPosts; i++ {
		if err := service.Add(makePost(i)); err != nil {
			t.Fatalf("Add returned an error: %s", err)
		}
	}

	t.Run("List all posts at once", func(t *testing.T) {
		expectedPosts := make([]types.Post, nPosts)

		for i := uint(0); i < nPosts; i++ {
			expectedPosts[i] = makePost(nPosts - i - 1)
		}

		cursor := listExpect(t, service, "", nPosts, expectedPosts, false)

		if cursor == "" {
			return
		}

		// List didn't return an empty cursor yet, but maybe the next call returns it...
		listExpect(t, service, cursor, nPosts, nil, true)
	})

	t.Run("Paginate", func(t *testing.T) {
		pageSize := uint(nPosts * 2 / 3) // so that we get empty cursor when listing the second page
		expectedPosts := make([]types.Post, pageSize)

		for i := uint(0); i < pageSize; i++ {
			expectedPosts[i] = makePost(nPosts - i - 1)
		}

		cursor := listExpect(t, service, "", pageSize, expectedPosts, false)

		if cursor == "" {
			t.Errorf("List for first page did return an empty cursor")
			return // not much else we can do...
		}

		nRemainingPosts := nPosts - pageSize
		expectedPosts = make([]types.Post, nRemainingPosts)

		for i := uint(0); i < nRemainingPosts; i++ {
			expectedPosts[i] = makePost(nRemainingPosts - i - 1)
		}

		cursor = listExpect(t, service, cursor, pageSize, expectedPosts, false)

		if cursor == "" {
			return
		}

		listExpect(t, service, cursor, pageSize, nil, true)
	})
}
