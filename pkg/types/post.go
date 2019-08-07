package types

import (
	"time"
)

type Post struct {
	ID      string    `json:"id"`
	Author  string    `json:"author"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
	Message string    `json:"message"`
}

func (p Post) Equal(other Post) bool {
	return p.ID == other.ID &&
		p.Author == other.Author &&
		p.Email == other.Email &&
		p.Created.Equal(other.Created) &&
		p.Message == other.Message
}
