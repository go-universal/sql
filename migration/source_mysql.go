package migration

import (
	"context"
	"database/sql"

	"github.com/go-universal/sql/mysql"
)

type mysqlSource struct {
	conn mysql.Connection
}

// NewMySQLSource creates a new MySQL migration source using the provided connection.
func NewMySQLSource(conn mysql.Connection) MigrationSource {
	return &mysqlSource{conn: conn}
}

func (ps *mysqlSource) Transaction(c context.Context, cb func(ExecutableScanner) error) error {
	tx, err := ps.conn.Database().BeginTx(c, nil)
	if err != nil {
		return err
	}

	if err := cb(&mysqlTX{tx: tx}); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (ps *mysqlSource) Exec(c context.Context, s string, args ...any) error {
	_, err := ps.conn.Database().ExecContext(c, s, args...)
	return err
}

func (ps *mysqlSource) Scan(c context.Context, s string, args ...any) (Rows, error) {
	rows, err := ps.conn.Database().QueryContext(c, s, args...)
	if err != nil {
		return nil, err
	}

	return &mysqlRows{rows: rows}, nil
}

// Implement ExecutableScanner for transaction
type mysqlTX struct {
	tx *sql.Tx
}

func (px *mysqlTX) Exec(c context.Context, s string, args ...any) error {
	_, err := px.tx.ExecContext(c, s, args...)
	return err
}

func (px *mysqlTX) Scan(c context.Context, s string, args ...any) (Rows, error) {
	rows, err := px.tx.QueryContext(c, s, args...)
	if err != nil {
		return nil, err
	}

	return &mysqlRows{rows: rows}, nil
}

// Implement Scanner row
type mysqlRows struct {
	rows *sql.Rows
}

func (ps *mysqlRows) Next() bool {
	return ps.rows.Next()
}

func (ps *mysqlRows) Scan(dest ...any) error {
	return ps.rows.Scan(dest...)
}

func (ps *mysqlRows) Close() {
	ps.rows.Close()
}
