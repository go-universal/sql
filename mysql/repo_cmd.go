package mysql

import (
	"context"
	"database/sql"
)

// Commander provides methods to construct and execute SQL commands.
type Commander interface {
	// Command sets the SQL query with '?' placeholders for parameters.
	Command(sql string) Commander

	// Replace substitutes a placeholder in the query string before execution.
	// Placeholders are in the @key format (e.g., "@sort", "@order").
	Replace(old, new string) Commander

	// Exec normalizes and executes the SQL command with the provided arguments.
	Exec(ctx context.Context, arguments ...any) (sql.Result, error)
}

type commander struct {
	db           Executable
	sql          string
	replacements []string
}

// NewCmd creates a new Commander instance with the provided Executable interface.
func NewCmd(e Executable) Commander {
	return &commander{
		db:           e,
		sql:          "",
		replacements: make([]string, 0),
	}
}

func (c *commander) Command(s string) Commander {
	c.sql = s
	return c
}

func (c *commander) Replace(o, n string) Commander {
	c.replacements = append(
		c.replacements,
		quoteField(o), n,
		o, n,
	)
	return c
}

func (c *commander) Exec(ctx context.Context, args ...any) (sql.Result, error) {
	if c.sql == "" {
		return nil, ErrEmptySQL
	}

	cmd := compile(c.sql, c.replacements...)
	return c.db.ExecContext(ctx, cmd, args...)
}
