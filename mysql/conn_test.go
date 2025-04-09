package mysql_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-universal/sql/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBuildDSN(t *testing.T) {
	config := mysql.NewConfig().
		Host("localhost").
		User("root").
		Password("root").
		Database("test")

	dsn := config.Build()
	expectedDSN := "root:root@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=true"
	assert.Equal(t, expectedDSN, dsn, "expected DSN does not match")
}

func TestNewConnection(t *testing.T) {
	ctx := context.Background()
	config := mysql.NewConfig().
		Host("localhost").
		User("root").
		Password("root").
		Database("test")

	conn, err := mysql.New(
		ctx, config.Build(),
		func(d *sql.DB) {
			d.SetMaxOpenConns(10)
		},
		func(d *sql.DB) {
			assert.Equal(t, 10, d.Stats().MaxOpenConnections, "connection modifier failed")
		},
	)
	require.NoError(t, err, "expected no error when creating the connection")
	defer conn.Close()
}

func TestConnectionManager(t *testing.T) {
	ctx := context.Background()
	config := mysql.NewConfig().
		Host("localhost").
		User("root").
		Password("root")

	manager := mysql.NewConnectionManager(config)
	defer manager.Close()

	err := manager.Connect(ctx, "test")
	require.NoError(t, err, "expected no error while connecting")

	_, exists := manager.Get("test")
	assert.True(t, exists, "expected connection to exist")

	err = manager.Remove("test")
	require.NoError(t, err, "expected no error while removing the connection")

	_, exists = manager.Get("test")
	assert.False(t, exists, "expected connection to not exist")
}

func TestTransaction(t *testing.T) {
	ctx := context.Background()
	config := mysql.NewConfig().
		Host("localhost").
		User("root").
		Password("root").
		Database("test")

	conn, err := mysql.New(ctx, config.Build())
	require.NoError(t, err, "expected no error when creating the connection")
	defer conn.Close()

	err = conn.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS test (id SERIAL PRIMARY KEY, name TEXT)")
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "testname")
		return err
	})
	require.NoError(t, err, "expected no error during transaction")
}
