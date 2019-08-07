package postservice

import (
	"github.com/pkg/errors"
)

// userError is used to mark errors caused by a wrong input from the user (as
// opposed to runtime errors).
type userError struct {
	err error
}

func (e *userError) Error() string {
	return e.err.Error()
}

// Type assertion to make sure we implement error correctly...
var _ error = &userError{}

// IsUserError checks if the given error is flagged as an error caused by a
// malformed user input or if it is a "normal" runtime error.
func IsUserError(e error) bool {
	return UserError(e) != nil
}

// UserError returns the user error inside the given error if any, or nil if e
// is nil or not a user error.
func UserError(e error) error {
	for e != nil {
		if uerror, ok := e.(*userError); ok {
			return uerror.err
		}

		cause := errors.Cause(e)

		if cause == e {
			// e didn't implement errors.causer
			return nil
		}

		e = cause
	}

	return nil
}
