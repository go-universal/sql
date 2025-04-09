package mysql

import (
	"context"
	"database/sql"
	"errors"
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
	// ExecContext runs a SQL command with optional parameters.
	// Returns a pgconn.CommandTag containing metadata about the execution result.
	ExecContext(ctx context.Context, sql string, args ...any) (sql.Result, error)
}

// Readable defines an interface for executing SQL queries.
type Readable interface {
	// QueryContext runs a SQL query with optional parameters and returns a pgx.Rows iterator.
	// The provided context is used for managing timeouts and cancellations.
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)

	// QueryRowContext runs a SQL query expecting a single result and returns a pgx.Row.
	// The provided context is used for managing timeouts and cancellations.
	QueryRowContext(ctx context.Context, sql string, args ...any) *sql.Row
}
