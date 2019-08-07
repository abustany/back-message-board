package poststore

import (
	"sync"

	avl "github.com/emirpasic/gods/trees/avltree"

	"github.com/abustany/back-message-board/pkg/types"
)

type memoryPostStore struct {
	sync.RWMutex
	posts       map[string]types.Post
	postsByDate *avl.Tree
}

// sortPostsByDateReverse compares two posts by their date, sorting the most
// recent first.
func sortPostsByDateReverse(a, b interface{}) int {
	aKey, bKey := a.(Cursor), b.(Cursor)

	if aKey.Created.Before(bKey.Created) {
		return 1
	}

	if aKey.Created.After(bKey.Created) {
		return -1
	}

	// Posts have the same creation time, sort by ID
	if aKey.ID < bKey.ID {
		return -1
	}

	if aKey.ID > bKey.ID {
		return 1
	}

	return 0
}

func NewMemoryPostStore() (Store, error) {
	return &memoryPostStore{
		posts:       map[string]types.Post{},
		postsByDate: avl.NewWith(sortPostsByDateReverse),
	}, nil
}

func (s *memoryPostStore) Add(post types.Post) error {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.posts[post.ID]; exists {
		return ErrIDAlreadyExists
	}

	s.posts[post.ID] = post
	s.postsByDate.Put(Cursor{ID: post.ID, Created: post.Created}, struct{}{})

	return nil
}

func (s *memoryPostStore) Update(post types.Post) error {
	s.Lock()
	defer s.Unlock()

	oldPost, exists := s.posts[post.ID]

	if !exists {
		return ErrIDNotFound
	}

	s.posts[post.ID] = post

	if oldPost.Created != post.Created {
		s.postsByDate.Remove(Cursor{ID: post.ID, Created: oldPost.Created})
		s.postsByDate.Put(Cursor{ID: post.ID, Created: post.Created}, struct{}{})
	}

	return nil
}

func (s *memoryPostStore) List(c Cursor, n uint) ([]types.Post, Cursor, error) {
	s.RLock()
	defer s.RUnlock()

	var node *avl.Node

	if c == EmptyCursor {
		node = s.postsByDate.Left()
	} else {
		node, _ = s.postsByDate.Ceiling(c)
	}

	if node == nil {
		// No more posts to iterate
		return nil, EmptyCursor, nil
	}

	posts := make([]types.Post, 0, n)

	for ; node != nil && uint(len(posts)) < n; node = node.Next() {
		posts = append(posts, s.posts[node.Key.(Cursor).ID])
	}

	var endCursor Cursor

	if node == nil {
		endCursor = EmptyCursor
	} else {
		endCursor = node.Key.(Cursor)
	}

	return posts, endCursor, nil
}
