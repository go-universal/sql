package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConfigModifier modifies database config before creation.
type ConfigModifier func(*pgxpool.Config)

// Connection represents a PostgreSQL database connection.
type Connection interface {
	// Database returns the underlying connection pool.
	Database() *pgxpool.Pool

	// Ping verifies the database connection by sending a simple query.
	Ping(ctx context.Context) error

	// Transaction executes a function within a transaction.
	// Commits if successful, rolls back on error.
	Transaction(ctx context.Context, cb func(pgx.Tx) error, options ...pgx.TxOptions) error

	// Close terminates the database connection pool.
	Close() error
}

type pgxConnection struct {
	db *pgxpool.Pool
}

// New creates a new PostgreSQL connection using a DSN string.
// Accepts optional ConfigModifier functions to adjust the configuration before establishing the connection.
func New(ctx context.Context, dsn string, modifiers ...ConfigModifier) (Connection, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	for _, modifier := range modifiers {
		modifier(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &pgxConnection{db: pool}, nil
}

func (d *pgxConnection) Database() *pgxpool.Pool {
	return d.db
}

func (d *pgxConnection) Ping(ctx context.Context) error {
	return d.db.Ping(ctx)
}

func (d *pgxConnection) Transaction(ctx context.Context, f func(pgx.Tx) error, opts ...pgx.TxOptions) error {
	tx, err := d.db.BeginTx(ctx, parseVariadic(pgx.TxOptions{}, opts...))
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (d *pgxConnection) Close() error {
	d.db.Close()
	return nil
}
