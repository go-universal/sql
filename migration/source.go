package migration

import "context"

// MigrationSource defines methods for running database migrations within a transaction.
type MigrationSource interface {
	// Transaction runs a function within a transaction context.
	// The transaction is committed if the function succeeds, or rolled back in case of an error.
	Transaction(ctx context.Context, callback func(ExecutableScanner) error) error

	// Exec executes a SQL command with the provided arguments.
	// Returns an error if the execution fails.
	Exec(ctx context.Context, sql string, arguments ...any) error

	// Scan executes a SQL query with the provided arguments and returns the result rows.
	// Returns an error if the query fails or if scanning the results encounters an issue.
	Scan(ctx context.Context, sql string, arguments ...any) (Rows, error)
}

// ExecutableScanner represents an entity capable of executing SQL commands and scanning results.
type ExecutableScanner interface {
	// Exec executes a SQL command with the provided arguments.
	// Returns an error if the execution fails.
	Exec(ctx context.Context, sql string, arguments ...any) error

	// Scan executes a SQL query with the provided arguments and returns the result rows.
	// Returns an error if the query fails or if scanning the results encounters an issue.
	Scan(ctx context.Context, sql string, arguments ...any) (Rows, error)
}

// Rows represents the set of results from a SQL query.
type Rows interface {
	// Next prepares the next row for reading.
	// Returns true if there is a next row, false if there are no more rows.
	Next() bool

	// Scan reads the current row's columns into the provided destination variables.
	// Returns an error if scanning the row fails.
	Scan(dest ...any) error

	// Close releases resources associated with the Rows.
	// It prevents further row enumeration after being called.
	Close()
}
