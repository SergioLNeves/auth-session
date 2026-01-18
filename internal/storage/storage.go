package storage

import "context"

type Storage interface {
	Ping(ctx context.Context) error
	Writer
	Reader
	Querier
}

type Writer interface {
	Insert(ctx context.Context, data any) error
}

type Reader interface {
	GetDB() any
}

type Querier interface {
	FindByEmail(ctx context.Context, email string, dest any) error
}
