package interfaces

import (
	"context"
)

type ResultSet interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

type CommandResult interface {
	RowsAffected() int64
}

type IDatabaseGateway interface {
	QueryRow(ctx context.Context, query string, dest any, args ...any) error
	Query(ctx context.Context, query string, args ...any) (ResultSet, error)
	Exec(ctx context.Context, query string, args ...any) error
	Close()
	RunMigrations() error
}
