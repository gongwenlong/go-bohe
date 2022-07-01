package transport

import "context"

type Server interface {
	GetType() string

	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
