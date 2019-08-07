package endpoint_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/abustany/back-message-board/pkg/endpoint"
	"github.com/abustany/back-message-board/pkg/postservice"
	"github.com/abustany/back-message-board/pkg/poststore"
	"github.com/abustany/back-message-board/pkg/types"
)

const adminUser = "admin"
const adminPassword = "r00tme"

func TestEndpoint(t *testing.T) {
	withUrl := func(f func(*testing.T, string)) func(*testing.T) {
		return func(t *testing.T) {
			//logger := log.NewNopLogger()
			logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
			store, err := poststore.NewMemoryPostStore()

			if err != nil {
				t.Fatalf("Error while creating store: %s", err)
			}

			adminUsers := map[string]string{
				adminUser: adminPassword,
			}

			ep := endpoint.NewHttpEndpoint(logger, postservice.New(store), adminUsers)
			server := httptest.NewServer(ep)
			defer server.Close()

			f(t, server.URL)
		}
	}

	t.Run("Add (invalid json)", withUrl(testAddInvalidJson))
	t.Run("Add", withUrl(testAdd))
	t.Run("Update", withUrl(testUpdate))
	t.Run("List (authentication)", withUrl(testListAuthentication))
	t.Run("List", withUrl(testList))
}

func testAddInvalidJson(t *testing.T, url string) {
	const invalidJson = "not json at all"

	res, err := http.Post(url+"/post", endpoint.JsonContentType, strings.NewReader(invalidJson))

	if err != nil {
		t.Fatalf("Error sending request: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("HTTP status when sending invalid request should be bad request")
	}
}

func postPost(t *testing.T, url string, post types.Post, auth bool, expectedStatus int) {
	buffer := bytes.Buffer{}

	if err := json.NewEncoder(&buffer).Encode(post); err != nil {
		t.Fatalf("Error while encoding JSON: %s", err)
	}

	req, err := http.NewRequest("POST", url, &buffer)

	if err != nil {
		t.Fatalf("Error while creating request: %s", err)
	}

	req.Header.Set("Content-Type", endpoint.JsonContentType)

	if auth {
		req.SetBasicAuth(adminUser, adminPassword)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("Error sending request: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != expectedStatus {
		t.Errorf("Unexpected HTTP status, got %d, expected %d", res.StatusCode, expectedStatus)
	}
}

func listPostsFull(t *testing.T, serverUrl, cursor string, pageSize int, expectedNumber int) ([]types.Post, string) {
	req, err := http.NewRequest("GET", serverUrl+"/admin/posts", nil)

	if err != nil {
		t.Fatalf("Error while creating request: %s", err)
	}

	queryParams := url.Values{}
	queryParams.Set("n", strconv.Itoa(pageSize))
	queryParams.Set("cursor", cursor)
	req.URL.RawQuery = queryParams.Encode()

	req.SetBasicAuth(adminUser, adminPassword)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("Error while sending request: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code for list response: %d", res.StatusCode)
	}

	var listResponse endpoint.ListResponse

	if err := json.NewDecoder(res.Body).Decode(&listResponse); err != nil {
		t.Fatalf("Decoding list response failed: %s", err)
	}

	if len(listResponse.Posts) != expectedNumber {
		t.Fatalf("List returned %d posts, expected %d", len(listResponse.Posts), expectedNumber)
	}

	return listResponse.Posts, listResponse.Next
}

func listPosts(t *testing.T, url string, expectedNumber int) []types.Post {
	posts, _ := listPostsFull(t, url, "", 100, expectedNumber)

	return posts
}

func testAdd(t *testing.T, url string) {
	post := types.Post{
		Author:  "John",
		Email:   "john@domain.com",
		Message: "this is my message",
	}

	postPost(t, url+"/post", post, false, http.StatusCreated)

	posts := listPosts(t, url, 1)

	posts[0].ID = ""
	posts[0].Created = time.Time{}

	if !posts[0].Equal(post) {
		t.Errorf("Unexpected post returned after adding: got %+v, expected %+v", posts[0], post)
	}
}

func testUpdate(t *testing.T, url string) {
	post := types.Post{
		Author:  "John",
		Email:   "john@domain.com",
		Message: "this is my message",
	}

	postPost(t, url+"/post", post, true, http.StatusCreated)

	posts := listPosts(t, url, 1)

	t.Run("Authentication", func(t *testing.T) {
		postPost(t, url+"/admin/posts", posts[0], false, http.StatusUnauthorized)
	})

	t.Run("Full update", func(t *testing.T) {
		oldPost := posts[0]
		oldPost.Message = "I changed my mind"

		postPost(t, url+"/admin/posts", oldPost, true, http.StatusOK)

		posts = listPosts(t, url, 1)

		if !posts[0].Equal(oldPost) {
			t.Errorf("Unexpected post after update: got %+v, expected %+v", posts[0], oldPost)
		}
	})

	t.Run("Partial update", func(t *testing.T) {
		oldPost := posts[0]
		oldPost.Message = "and again"
		patch := types.Post{ID: oldPost.ID, Message: oldPost.Message}

		postPost(t, url+"/admin/posts", patch, true, http.StatusOK)

		posts = listPosts(t, url, 1)

		if !posts[0].Equal(oldPost) {
			t.Errorf("Unexpected post after update: got %+v, expected %+v", posts[0], oldPost)
		}
	})
}

func testListAuthentication(t *testing.T, url string) {
	res, err := http.Get(url + "/admin/posts")

	if err != nil {
		t.Fatalf("Error while sending request: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Unexpected status code with unauthenticated list request: got %d, expected %d", res.StatusCode, http.StatusUnauthorized)
	}
}

func testList(t *testing.T, url string) {
	listPosts(t, url, 0)

	posts := []types.Post{
		{Author: "A1", Email: "E1", Message: "M1"},
		{Author: "A2", Email: "E2", Message: "M2"},
	}

	for _, p := range posts {
		postPost(t, url+"/post", p, false, http.StatusCreated)
	}

	list, cursor := listPostsFull(t, url, "", 1, 1)

	list[0].ID = ""
	list[0].Created = time.Time{}

	// Most recent posts first
	if !list[0].Equal(posts[1]) {
		t.Errorf("Unexpected post on first page: got %+v ,expected %+v", list[0], posts[1])
	}

	if cursor == "" {
		t.Errorf("List returned an empty cursor with some posts remaining to list")
	}

	list, cursor = listPostsFull(t, url, cursor, 1, 1)

	list[0].ID = ""
	list[0].Created = time.Time{}

	if !list[0].Equal(posts[0]) {
		t.Errorf("Unexpected post on first page: got %+v ,expected %+v", list[0], posts[0])
	}

	if cursor != "" {
		_, cursor = listPostsFull(t, url, cursor, 1, 0)
	}

	if cursor != "" {
		t.Errorf("List returned a non empty cursor at the end of the list")
	}
}
