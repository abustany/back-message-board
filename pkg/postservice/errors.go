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

func IsUserError(e error) bool {
	for e != nil {
		if _, ok := e.(*userError); ok {
			return true
		}

		cause := errors.Cause(e)

		if cause == e {
			// e didn't implement errors.causer
			return false
		}

		e = cause
	}

	return false
}
