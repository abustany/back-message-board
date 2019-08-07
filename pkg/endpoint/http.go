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
	"github.com/abustany/back-message-board/pkg/types"
)

type HttpEndpoint struct {
	router  *mux.Router
	service postservice.Service
}

type ListResponse struct {
	Posts []types.Post `json:"posts"`
	Next  string       `json:"next,omitempty"`
}

// Type assertion
var _ http.Handler = &HttpEndpoint{}

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

	adminRouter.Methods("GET").Path("/posts").Handler(adminHandler(http.HandlerFunc(endpoint.handleList)))
	adminRouter.Methods("POST").Path("/posts").Handler(adminHandler(WithPost(endpoint.handleEdit)))

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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (e *HttpEndpoint) handleEdit(post types.Post) (int, error) {
	err := e.service.Update(post)

	if err != nil {
		return 0, errors.Wrap(err, "Error while editing post")
	}

	return http.StatusOK, nil
}