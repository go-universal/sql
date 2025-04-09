package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Commonly used errors for database operations.
var (
	ErrEmptySQL   = errors.New("SQL command cannot be empty")
	ErrStructOnly = errors.New("expected type must be a struct")
)

// Transformer defines an interface for decoding and transforming data.
type Transformer interface {
	// Transform processes and extracts data.
	// Returns an error if the transformation fails.
	Transform() error
}

// Executable defines an interface for executing SQL commands.
type Executable interface {
	// Exec runs a SQL command with optional parameters.
	// Returns a pgconn.CommandTag containing metadata about the execution result.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Readable defines an interface for executing SQL queries.
type Readable interface {
	// Query runs a SQL query with optional parameters and returns a pgx.Rows iterator.
	// The provided context is used for managing timeouts and cancellations.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow runs a SQL query expecting a single result and returns a pgx.Row.
	// The provided context is used for managing timeouts and cancellations.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
