package application

import (
	"context"
	"io"
)

// Service - service implementation interface.
type Service interface {
	Serve() error
	io.Closer
}

// Constructor - template function for creating a service where dependencies are initialized.
type Constructor func(ctx context.Context, app *Application) (Service, error)
