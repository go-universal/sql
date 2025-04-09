package mysql_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-universal/sql/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	type User struct {
		Id   int    `db:"id"`
		Name string `db:"name"`
	}

	ctx := context.Background()
	config := mysql.NewConfig().
		Host("localhost").
		User("root").
		Password("root").
		Database("test")

	conn, err := mysql.New(ctx, config.Build())
	require.NoError(t, err, "expected no error when creating the connection")

	t.Run("Init", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx *sql.Tx) error {
			_, err := mysql.NewCmd(tx).
				Command("DROP TABLE IF EXISTS users CASCADE;").
				Exec(ctx)
			if err != nil {
				return err
			}

			_, err = mysql.NewCmd(tx).
				Command("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT);").
				Exec(ctx)
			return err
		})
		require.NoError(t, err, "expected no error during initialization")
	})

	t.Run("Insert", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx *sql.Tx) error {
			for idx, name := range []string{"John Doe", "Jack Ma"} {
				u := User{
					Id:   idx + 1,
					Name: name,
				}

				_, err := mysql.NewInserter[User](tx).
					Table("users").
					Insert(ctx, u, mysql.OnlyFields("name"))
				if err != nil {
					return err
				}
			}
			return nil
		})
		require.NoError(t, err, "expected no error during insert")
	})

	t.Run("Update", func(t *testing.T) {
		err := conn.Transaction(ctx, func(tx *sql.Tx) error {
			for idx, name := range []string{"John Doe New", "Jack Ma New"} {
				u := User{
					Id:   idx + 1,
					Name: name,
				}

				_, err := mysql.NewUpdater[User](tx).
					Table("users").
					Where("id = ?", u.Id).
					Update(ctx, u, mysql.SkipFields("id"))
				if err != nil {
					return err
				}
			}
			return nil
		})
		require.NoError(t, err, "expected no error during update")
	})

	t.Run("Count", func(t *testing.T) {
		count, err := mysql.NewCounter(conn.Database()).
			Query("SELECT COUNT(*) FROM users;").
			Count(ctx)
		require.NoError(t, err, "expected no error during count")
		assert.Equal(t, int64(2), count, "expected 2 users")
	})

	t.Run("Single", func(t *testing.T) {
		jack, err := mysql.NewFinder[User](conn.Database()).
			Query("SELECT * FROM users WHERE id = ?;").
			Struct(ctx, 2)

		require.NoError(t, err, "expected no error when finding user")
		require.NotNil(t, jack, "expected user to be found")
		assert.Equal(t, "Jack Ma New", jack.Name, `expected "Jack Ma New", got %s`, jack.Name)
	})

	t.Run("Multiple", func(t *testing.T) {
		users, err := mysql.NewFinder[User](conn.Database()).
			Query("SELECT * FROM users;").
			Structs(ctx)

		require.NoError(t, err, "expected no error when finding multiple users")
		assert.Len(t, users, 2, "expected 2 users, got %d", len(users))
	})
}
