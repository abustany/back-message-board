package poststore_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/abustany/back-message-board/pkg/poststore"
	"github.com/abustany/back-message-board/pkg/types"
)

func testStore(t *testing.T, storeFactory func() poststore.Store) {
	withStore := func(f func(t *testing.T, store poststore.Store)) func(*testing.T) {
		return func(t *testing.T) {
			f(t, storeFactory())
		}
	}

	t.Run("Add", withStore(testAdd))
	t.Run("Update", withStore(testUpdate))
	t.Run("List", withStore(testList))
}

func checkPosts(t *testing.T, store poststore.Store, expected []types.Post) {
	posts, _, err := store.List(poststore.EmptyCursor, 100)

	if err != nil {
		t.Errorf("List returned an error: %s", err)
		return
	}

	if len(posts) != len(expected) {
		t.Errorf("List returned %d posts, expected %d", len(posts), len(expected))
	}

	for i := range expected {
		if !posts[i].Equal(expected[i]) {
			t.Errorf("List returned an unexpected post at index %d: got %+v, expected %+v", i, posts[i], expected[i])
		}
	}
}

func testAdd(t *testing.T, store poststore.Store) {
	post := types.Post{
		ID: "ID",
	}

	if err := store.Add(post); err != nil {
		t.Errorf("Add returned an error when adding a new post: %s", err)
	}

	checkPosts(t, store, []types.Post{post})

	if err := store.Add(post); err == nil {
		t.Errorf("Add didn't return an error when adding a post with an existing ID")
	} else if err != poststore.ErrIDAlreadyExists {
		t.Errorf("Add returned an unexpected error when adding a post with an existing ID: %s", err)
	}

	checkPosts(t, store, []types.Post{post})
}

func testUpdate(t *testing.T, store poststore.Store) {
	post := types.Post{
		ID:      "ID",
		Author:  "Author1",
		Email:   "Email1",
		Created: time.Now(),
		Message: "Message1",
	}

	if err := store.Update(post); err == nil {
		t.Errorf("Update didn't return an error when updating a non existing post")
	} else if err != poststore.ErrIDNotFound {
		t.Errorf("Update returned an unexpected error when updating a non existing post: %s", err)
	}

	checkPosts(t, store, []types.Post{})

	if err := store.Add(post); err != nil {
		t.Fatalf("Add returned an error: %s", err)
	}

	checkPosts(t, store, []types.Post{post})

	post.Author = "Author2"
	post.Email = "Email2"
	post.Created = time.Now()
	post.Message = "Message2"

	if err := store.Update(post); err != nil {
		t.Errorf("Update returned an error when updating an existing post: %s", err)
	}

	checkPosts(t, store, []types.Post{post})

	post.Author = "Author3"

	if err := store.Update(types.Post{ID: post.ID, Author: post.Author}); err != nil {
		t.Errorf("Partial update returned an error when updating an existing post: %s", err)
	}

	checkPosts(t, store, []types.Post{post})
}

func testList(t *testing.T, store poststore.Store) {
	now := time.Now().Unix()
	const nPosts = 100

	makePost := func(idx int) types.Post {
		idxStr := fmt.Sprintf("%04d", idx)

		return types.Post{
			ID:      "ID" + idxStr,
			Author:  "Author " + idxStr,
			Email:   "Email " + idxStr,
			Created: time.Unix(now+int64(idx), 0),
			Message: "Hello " + idxStr,
		}
	}

	t.Run("Empty store", func(t *testing.T) {
		posts, cursor, err := store.List(poststore.EmptyCursor, 10)

		if err != nil {
			t.Errorf("List on an empty store returned an error: %s", err)
		}

		if len(posts) != 0 {
			t.Error("List on an empty store didn't return an empty list of posts")
		}

		if cursor != poststore.EmptyCursor {
			t.Error("List on an empty store didn't return an empty cursor")
		}

		for i := 0; i < nPosts; i++ {
			if err := store.Add(makePost(i)); err != nil {
				t.Fatalf("Add returned an error: %s", err)
			}
		}
	})

	t.Run("List all posts at once", func(t *testing.T) {
		posts, cursor, err := store.List(poststore.EmptyCursor, nPosts)

		if err != nil {
			t.Errorf("List returned an error: %s", err)
		}

		if len(posts) != nPosts {
			t.Errorf("List returned %d posts, expected %d", len(posts), nPosts)
		} else {
			for i := 0; i < nPosts; i++ {
				// Posts were added with an increasing creation time, but List should
				// return the most recent first
				expected := makePost(nPosts - i - 1)

				if !posts[i].Equal(expected) {
					t.Errorf("List returned an unexpected post at index %d: got %+v, expected %+v", i, posts[i], expected)
				}
			}
		}

		if cursor == poststore.EmptyCursor {
			return
		}

		// List didn't return EmptyCursor yet, but maybe the next call returns it...
		posts, cursor, err = store.List(cursor, nPosts)

		if err != nil {
			t.Errorf("List after first page returned an error: %s", err)
		}

		if len(posts) != 0 {
			t.Errorf("List after first page returned %d posts, expected 0", len(posts))
		}

		if cursor != poststore.EmptyCursor {
			t.Errorf("List didn't return an empty cursor")
		}
	})

	t.Run("Paginate", func(t *testing.T) {
		pageSize := uint(nPosts * 2 / 3) // so that we get empty cursor when listing the second page
		posts, cursor, err := store.List(poststore.EmptyCursor, pageSize)

		if err != nil {
			t.Errorf("List for first page returned an error: %s", err)
		}

		if uint(len(posts)) != pageSize {
			t.Errorf("List for first page returned %d posts, expected %d", len(posts), pageSize)
		} else {
			for i := 0; i < int(pageSize); i++ {
				expected := makePost(nPosts - i - 1)

				if !posts[i].Equal(expected) {
					t.Errorf("List for first page returned an unexpected post at index %d: got %+v, expected %+v", i, posts[i], expected)
				}
			}
		}

		if cursor == poststore.EmptyCursor {
			t.Errorf("List for first page did return an empty cursor")
			return // not much else we can do...
		}

		posts, cursor, err = store.List(cursor, pageSize)

		if err != nil {
			t.Errorf("List for second page returned an error: %s", err)
		}

		expectedNPosts := nPosts - pageSize

		if uint(len(posts)) != expectedNPosts {
			t.Errorf("List for second page returned %d posts, expected %d", len(posts), expectedNPosts)
		} else {
			for i := 0; i < int(expectedNPosts); i++ {
				expected := makePost(nPosts - int(pageSize) - i - 1)

				if !posts[i].Equal(expected) {
					t.Errorf("List for first page returned an unexpected post at index %d: got %+v, expected %+v", i, posts[i], expected)
				}
			}
		}
	})
}
