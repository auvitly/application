package application

import (
	"context"
	"io"
)

type Service interface {
	Serve() error
	io.Closer
}

type Constructor func(ctx context.Context) (Service, error)
