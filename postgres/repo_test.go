package postgres_test

import (
	"context"
	"testing"

	"github.com/go-universal/sql/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	type User struct {
		Id   int    `db:"id"`
		Name string `db:"name"`
	}

	ctx := context.Background()
	config := postgres.NewConfig().
		Host("localhost").
		Port(5432).
		User("postgres").
		Password("root").
		Database("test")

	conn, err := postgres.New(ctx, config.Build())
	require.NoError(t, err, "expected no error when creating connection")
	defer conn.Close()

	t.Run("Init", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx pgx.Tx) error {
			_, err := postgres.NewCmd(tx).
				Command("DROP TABLE IF EXISTS users CASCADE;").
				Exec(ctx)
			if err != nil {
				return err
			}

			_, err = postgres.NewCmd(tx).
				Command("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT);").
				Exec(ctx)
			return err
		})
		require.NoError(t, err, "expected no error during table initialization")
	})

	t.Run("Insert", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx pgx.Tx) error {
			for idx, name := range []string{"John Doe", "Jack Ma"} {
				u := User{
					Id:   idx + 1,
					Name: name,
				}

				_, err = postgres.NewInserter[User](tx).
					Table("users").
					Insert(ctx, u, postgres.OnlyFields("name"))
				if err != nil {
					return err
				}
			}
			return nil
		})
		require.NoError(t, err, "expected no error during insertion")
	})

	t.Run("Update", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx pgx.Tx) error {
			for idx, name := range []string{"John Doe New", "Jack Ma New"} {
				u := User{
					Id:   idx + 1,
					Name: name,
				}

				_, err = postgres.NewUpdater[User](tx).
					Table("users").
					Where("id = ?", u.Id).
					Update(ctx, u, postgres.SkipFields("id"))
				if err != nil {
					return err
				}
			}
			return nil
		})
		require.NoError(t, err, "expected no error during update")
	})

	t.Run("Count", func(t *testing.T) {
		count, err := postgres.NewCounter(conn.Database()).
			Query("SELECT COUNT(*) FROM users;").
			Count(ctx)
		require.NoError(t, err, "expected no error during count query")
		assert.Equal(t, int64(2), count, "expected 2 users, got a different count")
	})

	t.Run("Single", func(t *testing.T) {
		jack, err := postgres.NewFinder[User](conn.Database()).
			Query("SELECT * FROM users WHERE id = ?;").
			Struct(ctx, 2)

		require.NoError(t, err, "expected no error during finding single user")
		require.NotNil(t, jack, "expected user, got nil")
		assert.Equal(t, "Jack Ma New", jack.Name, `expected "Jack Ma New", got different name`)
	})

	t.Run("Multiple", func(t *testing.T) {
		users, err := postgres.NewFinder[User](conn.Database()).
			Query("SELECT * FROM users;").
			Structs(ctx)

		require.NoError(t, err, "expected no error during finding multiple users")
		assert.Len(t, users, 2, "expected 2 users, got a different number")
	})
}
