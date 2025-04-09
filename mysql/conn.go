package mysql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// ConfigModifier modifies database config before creation.
type ConfigModifier func(*sql.DB)

// Connection represents a MySQL database connection.
type Connection interface {
	// Database returns the underlying connection pool.
	Database() *sql.DB

	// Ping verifies the database connection by sending a simple query.
	Ping(ctx context.Context) error

	// Transaction executes a function within a transaction.
	// Commits if successful, rolls back on error.
	Transaction(ctx context.Context, cb func(*sql.Tx) error) error

	// Close terminates the database connection pool.
	Close() error
}

type mysqlConnection struct {
	db *sql.DB
}

// New creates a new MySQL connection using a DSN string.
// Accepts optional ConfigModifier functions to adjust the configuration before establishing the connection.
func New(ctx context.Context, dsn string, modifiers ...ConfigModifier) (Connection, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	for _, modifier := range modifiers {
		modifier(db)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &mysqlConnection{db: db}, nil
}

func (d *mysqlConnection) Database() *sql.DB {
	return d.db
}

func (d *mysqlConnection) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *mysqlConnection) Transaction(ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *mysqlConnection) Close() error {
	return d.db.Close()
}
