package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Updater provides methods for updating records in a specified database table.
type Updater[T any] interface {
	// Table sets the target table for the update operation.
	Table(table string) Updater[T]

	// Where specifies the WHERE condition for the update.
	Where(condition string, args ...any) Updater[T]

	// Update performs the update operation with the provided record and options.
	Update(ctx context.Context, record T, options ...RepositoryOption) (sql.Result, error)
}

type updater[T any] struct {
	db     Executable
	table  string
	where  string
	args   []any
	option *options
}

// NewUpdater initializes and returns a new Updater instance for the given Executable interface.
func NewUpdater[T any](e Executable) Updater[T] {
	return &updater[T]{
		db: e,
		option: &options{
			only:    make([]string, 0),
			exclude: make([]string, 0),
		},
	}
}

func (u *updater[T]) Table(t string) Updater[T] {
	u.table = t
	return u
}

func (u *updater[T]) Where(w string, args ...any) Updater[T] {
	u.where = w
	u.args = append([]any{}, args...)
	return u
}

func (u *updater[T]) Update(ctx context.Context, v T, options ...RepositoryOption) (sql.Result, error) {
	if !isStruct[T]() {
		return nil, ErrStructOnly
	}

	if u.table == "" || u.where == "" {
		return nil, ErrEmptySQL
	}

	for _, opt := range options {
		opt(u.option)
	}

	columns := structColumns(v, u.option.only, u.option.exclude)
	values := structValues(v, u.option.only, u.option.exclude)
	values = append(values, u.args...)

	for i, col := range columns {
		columns[i] = fmt.Sprintf("%s = ?", col)
	}

	cmd := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s;",
		u.table,
		strings.Join(columns, ","),
		u.where,
	)
	return u.db.ExecContext(ctx, cmd, values...)
}
