package query_test

import (
	"errors"
	"io/fs"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-universal/sql/query"
)

type MockFS struct {
	files map[string]string
}

func (*MockFS) Exists(path string) (bool, error)                        { return false, nil }
func (*MockFS) Open(path string) (fs.File, error)                       { return nil, nil }
func (*MockFS) Search(dir, phrase, ignore, ext string) (*string, error) { return nil, nil }
func (*MockFS) Find(dir, pattern string) (*string, error)               { return nil, nil }
func (*MockFS) FS() fs.FS                                               { return nil }
func (*MockFS) Http() http.FileSystem                                   { return nil }
func (f *MockFS) Lookup(dir, pattern string) ([]string, error) {
	names := make([]string, 0)
	for k := range f.files {
		names = append(names, k)
	}
	return names, nil
}
func (f *MockFS) ReadFile(path string) ([]byte, error) {
	v, ok := f.files[path]
	if !ok {
		return nil, errors.New("file not found")
	}
	return []byte(v), nil
}

func TestQuery(t *testing.T) {
	fs := &MockFS{
		files: map[string]string{
			"database/queries/user.sql": `
-- { query: list }
SELECT * FROM users WHERE deleted_at IS NULL;


-- { undefined: unsupported }
SELECT name, family
FROM users
WHERE
	deleted_at IS NULL
	AND age > 18
	AND name ILIKE '%@phrase%';

-- { query: single }
SELECT id, name, age FROM users WHERE @conditions;
			`,
		},
	}

	manager, err := query.NewQueryManager(fs, query.WithRoot("database/queries"))
	require.NoError(t, err)

	t.Run("Unsupported query should be ignored", func(t *testing.T) {
		q := manager.Get("user/unsupported")
		assert.Empty(t, q, "unsupported section must be ignored")
	})

	t.Run("Non-existent query should return false", func(t *testing.T) {
		_, exists := manager.Find("not-exists-query")
		assert.False(t, exists, "not-exists-query should not exist")
	})

	t.Run("Should get expected 'list' query", func(t *testing.T) {
		expected := `SELECT * FROM users WHERE deleted_at IS NULL;`
		q := manager.Get("user/list")
		assert.Equal(t, expected, q)
	})

	t.Run("Should build expected 'single' query", func(t *testing.T) {
		expected := `SELECT id, name, age FROM users WHERE deleted_at IS NULL AND (name = ? OR family = ?);`
		q := manager.Query("user/single").
			And("deleted_at IS NULL").
			AndClosure("name = ? OR family = ?").
			Build()
		assert.Equal(t, expected, q)
	})
}
