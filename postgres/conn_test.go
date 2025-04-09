package postgres_test

import (
	"context"
	"testing"

	"github.com/go-universal/sql/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBuildDSN(t *testing.T) {
	config := postgres.NewConfig().
		Host("localhost").
		Port(5432).
		User("postgres").
		Password("password").
		Database("test").
		SSLMode("disable").
		MaxConns(7).
		MinConns(2)

	dsn := config.Build()
	expectedDSN := "postgres://postgres:password@localhost:5432/test?pool_max_conns=7&pool_min_conns=2&sslmode=disable"
	assert.Equal(t, expectedDSN, dsn, "Expected DSN does not match the built DSN")
}

func TestNewConnection(t *testing.T) {
	ctx := context.Background()
	config := postgres.NewConfig().
		Host("localhost").
		Port(5432).
		User("postgres").
		Password("root").
		Database("postgres")

	conn, err := postgres.New(
		ctx, config.Build(),
		func(c *pgxpool.Config) { c.MaxConns = 7 },
		func(c *pgxpool.Config) {
			assert.Equal(t, int32(7), c.MaxConns, "connection modifier failed to set MaxConns")
		},
	)
	require.NoError(t, err, "Expected no error when creating new connection")
	defer conn.Close()
}

func TestConnectionManager(t *testing.T) {
	ctx := context.Background()
	config := postgres.NewConfig().
		Host("localhost").
		Port(5432).
		User("postgres").
		Password("root").
		Database("test").
		SSLMode("disable")

	manager := postgres.NewConnectionManager(config)
	defer manager.Close()

	err := manager.Connect(ctx, "test")
	require.NoError(t, err, "Expected no error when connecting to manager")

	_, exists := manager.Get("test")
	require.True(t, exists, "Expected connection to exist after being connected")

	err = manager.Remove("test")
	require.NoError(t, err, "Expected no error when removing connection")

	_, exists = manager.Get("test")
	require.False(t, exists, "Expected connection to not exist after being removed")
}

func TestTransaction(t *testing.T) {
	ctx := context.Background()
	config := postgres.NewConfig().
		Host("localhost").
		Port(5432).
		User("postgres").
		Password("root").
		Database("test").
		SSLMode("disable")

	conn, err := postgres.New(ctx, config.Build())
	require.NoError(t, err, "Expected no error when creating new connection")
	defer conn.Close()

	err = conn.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS test (id SERIAL PRIMARY KEY, name TEXT)")
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "INSERT INTO test (name) VALUES ($1)", "testname")
		return err
	})
	require.NoError(t, err, "Expected no error when performing transaction")
}
