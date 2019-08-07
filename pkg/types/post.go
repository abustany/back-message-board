// Package types gather the common types shared across various packages
package types

import (
	"time"
)

// Post describes a post in the message board.
type Post struct {
	// Unique ID of the post
	ID string `json:"id"`
	// Post author name
	Author string `json:"author"`
	// Post author email
	Email string `json:"email"`
	// Post creation time
	Created time.Time `json:"created"`
	// Post contents
	Message string `json:"message"`
}

// Equal returns true if and only if the posts p and other are equal. Comparing
// two posts using == is not always safe because of the Created field, for the
// same reason that comparing two time.Time values using == is not always safe.
func (p Post) Equal(other Post) bool {
	return p.ID == other.ID &&
		p.Author == other.Author &&
		p.Email == other.Email &&
		p.Created.Equal(other.Created) &&
		p.Message == other.Message
}
