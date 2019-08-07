package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/abustany/back-message-board/pkg/postservice"
	"github.com/abustany/back-message-board/pkg/types"
)

// JsonContentType is the MIME type of JSON requests/responses
const JsonContentType = "application/json"

type capturingResponseWriter struct {
	w    http.ResponseWriter
	code int
}

var _ http.ResponseWriter = &capturingResponseWriter{}

func (c *capturingResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *capturingResponseWriter) Write(data []byte) (int, error) {
	if c.code == 0 {
		c.code = http.StatusOK
	}

	return c.w.Write(data)
}

func (c *capturingResponseWriter) WriteHeader(statusCode int) {
	if c.code == 0 {
		c.code = statusCode
	}

	c.w.WriteHeader(statusCode)
}

// WithLogging wraps a http.Handler, writing a log message to the given logger
// at the end of each request with the URL, returned status code, elapsed time
// etc.
func WithLogging(logger log.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := capturingResponseWriter{w: w}

		defer func(start time.Time) {
			logger.Log(
				"event", "api_request",
				"method", r.Method,
				"url", r.URL.String(),
				"status", writer.code,
				"elapsed", time.Since(start),
			)
		}(time.Now())

		handler.ServeHTTP(&writer, r)
	})
}

// WithContentType wraps a http.Handler, rejecting requests that don't have a
// JSON content type.
func WithContentType(contentType string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != JsonContentType {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Invalid content type")
		}

		handler.ServeHTTP(w, r)
	})
}

// WriteError write the given error to the ResponseWriter, using the
// appropriate HTTP status code depending on whether the error is a user or an
// internal error.
func WriteError(w http.ResponseWriter, err error) {
	if userError := postservice.UserError(err); userError != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, userError.Error())
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

// WithPost adapts an http.Handler to a function handling an HTTP request where
// the request body is a single types.Post object. The error returned by the
// function is written back to the response using WriteError.
func WithPost(do func(post types.Post) (int, error)) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		post := types.Post{}

		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Malformed JSON input")
			return
		}

		statusCode, err := do(post)

		if err != nil {
			WriteError(w, err)
		} else {
			w.WriteHeader(statusCode)
		}
	}

	return WithContentType(JsonContentType, http.HandlerFunc(handler))
}

// RequestAuthenticator is a common interface to all HTTP request authentication
// functions.
type RequestAuthenticator interface {
	// Authenticate returns true if and only if the request has valid
	// credentials.
	Authenticate(r *http.Request) (bool, error)
}

// WithAuthentication wraps an http.Handler, rejecting requests that don't get a
// valid result from the given authenticator.
func WithAuthentication(authenticator RequestAuthenticator, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, err := authenticator.Authenticate(r)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// BasicAuthenticator uses HTTP Basic Auth to authenticate requests
type BasicAuthenticator struct {
	// Users maps usernames to password
	//
	// Because BasicAuthenticator does not implement any locking, this map
	// shouldn't be modified once the HTTP handler starts handling requests.
	Users map[string]string
}

func (a *BasicAuthenticator) Authenticate(r *http.Request) (bool, error) {
	username, password, ok := r.BasicAuth()

	if !ok {
		return false, nil
	}

	if realPassword, knownUser := a.Users[username]; !knownUser || password != realPassword {
		return false, nil
	}

	return true, nil
}
