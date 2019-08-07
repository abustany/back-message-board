package poststore_test

import (
	"testing"

	"github.com/abustany/back-message-board/pkg/poststore"
)

func TestMemoryStore(t *testing.T) {
	testStore(t, func() poststore.Store {
		store, err := poststore.NewMemoryPostStore()

		if err != nil {
			t.Fatalf("NewMemoryPostStore returned an error: %s", err)
		}

		return store
	})
}
