package query

import (
	"regexp"
	"sync"

	"github.com/go-universal/fs"
)

// QueryManager defines methods for managing and retrieving SQL queries.
// Queries should be defined with `-- { query: name }` comments in the SQL files.
type QueryManager interface {
	// Load retrieves all queries from the filesystem and caches them.
	Load() error

	// Get retrieves the SQL query by name.
	// Queries are resolved by `relative/path/to/root/query` key.
	Get(name string) string

	// Find attempts to retrieve the SQL query by name and returns whether it was found.
	Find(name string) (string, bool)

	// Query builds a QueryBuilder for the specified query.
	Query(name string) QueryBuilder
}

type queryManager struct {
	root     string
	ext      string
	dev      bool
	fs       fs.FlexibleFS
	queries  map[string]string
	resolver PlaceholderResolver
	quote    QuoteResolver
	mutex    sync.RWMutex
}

// NewQueryManager initializes a query manager with the specified filesystem and options.
func NewQueryManager(fs fs.FlexibleFS, options ...Options) (QueryManager, error) {
	q := &queryManager{
		root:     ".",
		ext:      ".sql",
		dev:      false,
		fs:       fs,
		queries:  make(map[string]string),
		resolver: nil,
		quote:    nil,
	}

	for _, opt := range options {
		opt(q)
	}

	if err := q.Load(); err != nil {
		return nil, err
	}

	return q, nil
}

func (m *queryManager) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Locate files matching the specified extension
	files, err := m.fs.Lookup(m.root, ".*"+regexp.QuoteMeta(m.ext))
	if err != nil {
		return err
	}

	// Parse and cache queries from each file
	for _, file := range files {
		content, err := m.fs.ReadFile(file)
		if err != nil {
			return err
		}

		fName := toName(file, m.root, m.ext)
		queries, err := parseQueries(string(content))
		if err != nil {
			return err
		}

		// Store parsed queries with path-based keys for uniqueness
		for qName, query := range queries {
			m.queries[fName+"/"+qName] = query
		}
	}

	return nil
}

func (m *queryManager) Get(n string) string {
	if m.dev {
		m.Load()
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.queries[n]
}

func (m *queryManager) Find(n string) (string, bool) {
	if m.dev {
		m.Load()
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.queries[n]
	return v, ok
}

func (m *queryManager) Query(n string) QueryBuilder {
	return &queryBuilder{
		sql:          m.Get(n),
		resolver:     m.resolver,
		quote:        m.quote,
		conditions:   make([]queryItem, 0),
		replacements: make([]string, 0),
	}
}
