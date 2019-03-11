package database

import (
	"io"
	"time"
)

// Modeler defines the behaviour of all elements that can be stored in the
// Datastore
type Modeler interface {
	encode() (io.Reader, error)
}

type model struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
