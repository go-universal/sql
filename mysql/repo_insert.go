package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Inserter provides methods for inserting a struct into a specified database table.
type Inserter[T any] interface {
	// Table sets the target table for the insert operation.
	Table(table string) Inserter[T]

	// Insert performs an insert operation with the given record and options.
	Insert(ctx context.Context, record T, options ...RepositoryOption) (sql.Result, error)
}

type inserter[T any] struct {
	db     Executable
	table  string
	option *options
}

// NewInserter initializes and returns a new Inserter instance for the given Executable interface.
func NewInserter[T any](e Executable) Inserter[T] {
	return &inserter[T]{
		db: e,
		option: &options{
			only:    make([]string, 0),
			exclude: make([]string, 0),
		},
	}
}

func (i *inserter[T]) Table(t string) Inserter[T] {
	i.table = t
	return i
}

func (i *inserter[T]) Insert(ctx context.Context, v T, options ...RepositoryOption) (sql.Result, error) {
	if !isStruct[T]() {
		return nil, ErrStructOnly
	}

	if i.table == "" {
		return nil, ErrEmptySQL
	}

	for _, opt := range options {
		opt(i.option)
	}

	columns := structColumns(v, i.option.only, i.option.exclude)
	values := structValues(v, i.option.only, i.option.exclude)
	placeholders := make([]string, 0, len(columns))

	for range columns {
		placeholders = append(placeholders, "?")
	}

	cmd := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s);",
		i.table,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
	)
	return i.db.ExecContext(ctx, cmd, values...)
}
