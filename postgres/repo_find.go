package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Finder provides methods to construct and execute SQL queries that return structured results.
type Finder[T any] interface {
	// Query sets the SQL query string to be executed.
	Query(sql string) Finder[T]

	// Replace updates specific placeholders in the SQL query.
	// Placeholders are in the @key format (e.g., "@column_name").
	Replace(old, new string) Finder[T]

	// WithTransformer adds a transformation function to modify the result.
	WithTransformer(func(*T) error) Finder[T]

	// Rows executes the query and returns a pgx.Rows iterator for processing result rows.
	Rows(ctx context.Context, args ...any) (pgx.Rows, error)

	// Struct executes the query and retrieves a single result, or an error if the query fails.
	Struct(ctx context.Context, args ...any) (*T, error)

	// Structs executes the query and retrieves multiple results, or an error if the query fails.
	Structs(ctx context.Context, args ...any) ([]T, error)
}

type finder[T any] struct {
	db           Readable
	sql          string
	replacements []string
	transformers []func(*T) error
}

// NewFinder creates a new Finder instance with the provided Readable interface.
func NewFinder[T any](r Readable) Finder[T] {
	return &finder[T]{
		db:           r,
		sql:          "",
		replacements: make([]string, 0),
		transformers: make([]func(*T) error, 0),
	}
}

func (f *finder[T]) Query(s string) Finder[T] {
	f.sql = s
	return f
}

func (f *finder[T]) Replace(o, n string) Finder[T] {
	f.replacements = append(f.replacements, o, n)
	return f
}

func (f *finder[T]) WithTransformer(t func(*T) error) Finder[T] {
	f.transformers = append(f.transformers, t)
	return f
}

func (f *finder[T]) Rows(ctx context.Context, args ...any) (pgx.Rows, error) {
	if f.sql == "" {
		return nil, ErrEmptySQL
	}

	sql := compile(f.sql, f.replacements...)
	rows, err := f.db.Query(ctx, sql, args...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return rows, nil
}

func (f *finder[T]) Struct(ctx context.Context, args ...any) (*T, error) {
	if !isStruct[T]() {
		return nil, ErrStructOnly
	}

	rows, err := f.Rows(ctx, args...)
	if err != nil {
		return nil, err
	} else if rows == nil {
		return nil, nil
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	if tr, ok := any(&result).(Transformer); ok {
		if err := tr.Transform(); err != nil {
			return nil, err
		}
	}

	for _, transformer := range f.transformers {
		if err := transformer(&result); err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func (f *finder[T]) Structs(ctx context.Context, args ...any) ([]T, error) {
	if !isStruct[T]() {
		return nil, ErrStructOnly
	}

	rows, err := f.Rows(ctx, args...)
	if err != nil {
		return nil, err
	} else if rows == nil {
		return []T{}, nil
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	for i := range results {
		if tr, ok := any(&results[i]).(Transformer); ok {
			if err := tr.Transform(); err != nil {
				return nil, err
			}
		}

		for _, transformer := range f.transformers {
			if err := transformer(&results[i]); err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}
