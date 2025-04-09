package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// Inserter provides methods for inserting a struct into a specified database table.
type Inserter[T any] interface {
	// Table sets the target table for the insert operation.
	Table(table string) Inserter[T]

	// Insert performs an insert operation with the given record and options.
	Insert(ctx context.Context, record T, options ...RepositoryOption) (pgconn.CommandTag, error)
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

func (i *inserter[T]) Insert(ctx context.Context, v T, options ...RepositoryOption) (pgconn.CommandTag, error) {
	if !isStruct[T]() {
		return pgconn.CommandTag{}, ErrStructOnly
	}

	if i.table == "" {
		return pgconn.CommandTag{}, ErrEmptySQL
	}

	for _, opt := range options {
		opt(i.option)
	}

	columns := structColumns(v, i.option.only, i.option.exclude)
	values := structValues(v, i.option.only, i.option.exclude)
	placeholders := make([]string, 0, len(columns))

	for idx := range columns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
	}

	sql := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s);`,
		i.table,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
	)

	return i.db.Exec(ctx, sql, values...)
}
