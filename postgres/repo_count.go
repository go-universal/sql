package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Counter provides methods for constructing and executing SQL queries to count results.
type Counter interface {
	// Query sets the SQL query for counting rows.
	Query(sql string) Counter

	// Replace substitutes placeholders in the query string before execution.
	// Placeholders are in the @key format (e.g., "@column_name").
	Replace(old, new string) Counter

	// Count executes the query and returns the row count.
	// It uses the provided arguments for parameterized queries.
	// Returns the count and any errors encountered.
	Count(ctx context.Context, args ...any) (int64, error)
}

type counter struct {
	db           Readable
	sql          string
	replacements []string
}

// NewCounter creates a new Counter instance with the provided Readable interface.
func NewCounter(r Readable) Counter {
	return &counter{
		db:           r,
		sql:          "",
		replacements: make([]string, 0),
	}
}

func (c *counter) Query(s string) Counter {
	c.sql = s
	return c
}

func (c *counter) Replace(o, n string) Counter {
	c.replacements = append(c.replacements, o, n)
	return c
}

func (c *counter) Count(ctx context.Context, args ...any) (int64, error) {
	if c.sql == "" {
		return 0, ErrEmptySQL
	}

	var count int64
	sql := compile(c.sql, c.replacements...)
	err := c.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	return count, nil
}
