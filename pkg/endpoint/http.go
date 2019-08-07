// Package endpoint implements an HTTP endpoint to the message board service
package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/postservice"
	"github.com/abustany/back-message-board/pkg/poststore"
	"github.com/abustany/back-message-board/pkg/types"
)

// HttpEndpoint exposes the functionality of postervice.Service over HTTP
type HttpEndpoint struct {
	router  *mux.Router
	service postservice.Service
}

// ListResponse is the shape of List replies.
type ListResponse struct {
	// List of posts on that result page
	Posts []types.Post `json:"posts"`
	// Cursor to the next result page
	Next string `json:"next,omitempty"`
}

// Type assertion
var _ http.Handler = &HttpEndpoint{}

// NewHttpEndpoint returns a new instance of HttpEndpoint backed by the given
// service.
//
// HTTP requests will be logged to the given logger, and adminUsers will be used
// to authenticate users accessing the admin API. It should be a map where keys
// are usernames, and values the password for each username.
func NewHttpEndpoint(logger log.Logger, service postservice.Service, adminUsers map[string]string) *HttpEndpoint {
	endpoint := &HttpEndpoint{
		router:  mux.NewRouter(),
		service: service,
	}

	logger = log.With(logger, "module", "http")

	endpoint.router.Methods("POST").Path("/post").Handler(WithLogging(logger, WithPost(endpoint.handlePost)))

	adminRouter := endpoint.router.PathPrefix("/admin").Subrouter()

	authenticator := BasicAuthenticator{
		Users: adminUsers,
	}

	adminHandler := func(handler http.Handler) http.Handler {
		return WithLogging(logger, WithAuthentication(&authenticator, handler))
	}

	adminRouter.Methods("GET").Path("/posts/{id}").Handler(adminHandler(http.HandlerFunc(endpoint.handleGet)))
	adminRouter.Methods("GET").Path("/posts").Handler(adminHandler(http.HandlerFunc(endpoint.handleList)))
	adminRouter.Methods("POST").Path("/posts").Handler(adminHandler(WithPost(endpoint.handleEdit)))

	endpoint.router.Methods("GET").Path("/health").Handler(WithLogging(logger, http.HandlerFunc(endpoint.handleHealth)))

	return endpoint
}

func (e *HttpEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func (e *HttpEndpoint) handlePost(post types.Post) (int, error) {
	err := e.service.Add(post)

	if err != nil {
		return 0, errors.Wrap(err, "Error while adding post")
	}

	return http.StatusCreated, nil
}

func (e *HttpEndpoint) handleList(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	cursor := params.Get("cursor")
	pageSizeStr := params.Get("n")

	if pageSizeStr == "" {
		pageSizeStr = "0"
	}

	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid page size")
		return
	}

	posts, next, err := e.service.List(cursor, uint(pageSize))

	if err != nil {
		WriteError(w, err)
		return
	}

	response := ListResponse{
		Posts: posts,
		Next:  next,
	}

	w.Header().Set("Content-Type", JsonContentType)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (e *HttpEndpoint) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["id"]

	if postId == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	post, err := e.service.Get(postId)

	if postservice.UserError(err) == poststore.ErrIDNotFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", JsonContentType)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&post)
}

func (e *HttpEndpoint) handleEdit(post types.Post) (int, error) {
	err := e.service.Update(post)

	if err != nil {
		return 0, errors.Wrap(err, "Error while editing post")
	}

	return http.StatusOK, nil
}

func (e *HttpEndpoint) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
