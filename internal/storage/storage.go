package storage

import "context"

type Storage interface {
	Ping(ctx context.Context) error
	Writer
	Reader
	Querier
}

type Writer interface {
	Insert(ctx context.Context, table string, data any) error
	Update(ctx context.Context, table string, data any) error
	FindOneAndDelete(ctx context.Context, table string, id any, dest any) error
}

type Reader interface {
	GetDB() any
}

type Querier interface {
	FindByEmail(ctx context.Context, table, email string, dest any) error
	FindByID(ctx context.Context, table string, id any, dest any) error
}
