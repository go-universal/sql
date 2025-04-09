package migration

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/go-universal/fs"
)

// Migration defines the interface for managing database migrations.
// It includes methods for loading migration stages, initializing the migration table,
// retrieving summaries, and applying or rolling back migration stages.
type Migration interface {
	// Load loads migration stages from the filesystem and caches them.
	Load() error

	// Root returns the root directory of migration files.
	Root() string

	// Extension returns the file extension for migration files.
	Extension() string

	// IsDev indicates if it is in development mode.
	IsDev() bool

	// Initialize sets up the database migration table.
	Initialize() error

	// Summary returns an overview of the migration.
	Summary() (Summary, error)

	// Up applies migration stages.
	Up(stages []string, options ...MigrationOption) (Summary, error)

	// Down rolls back migration stages.
	Down(stages []string, options ...MigrationOption) (Summary, error)

	// Refresh rolls back and reapplies migration stages.
	Refresh(stages []string, options ...MigrationOption) (Summary, error)
}

type migration struct {
	root  string
	ext   string
	dev   bool
	files sortableFiles
	fs    fs.FlexibleFS
	db    MigrationSource
	mutex sync.RWMutex
}

// NewMigration initializes a migration with the specified database source, filesystem, and options.
func NewMigration(db MigrationSource, fs fs.FlexibleFS, options ...Option) (Migration, error) {
	mig := &migration{
		root:  ".",
		ext:   "sql",
		dev:   false,
		files: make(sortableFiles, 0),
		fs:    fs,
		db:    db,
	}

	for _, opt := range options {
		opt(mig)
	}

	if err := mig.Load(); err != nil {
		return nil, err
	}

	if err := mig.Initialize(); err != nil {
		return nil, err
	}

	return mig, nil
}

func (m *migration) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Locate files matching the specified extension
	files, err := m.fs.Lookup(m.root, `.*\.`+regexp.QuoteMeta(m.ext))
	if err != nil {
		return err
	}

	// Parse and cache migration stages from each file
	m.files = make(sortableFiles, 0)
	for _, file := range files {
		content, err := m.fs.ReadFile(file)
		if err != nil {
			return err
		}

		file := newMigrationFile(file, string(content))
		if file != nil {
			m.files = append(m.files, *file)
		}
	}

	sort.Sort(m.files)
	return nil
}

func (m *migration) Root() string {
	return m.root
}

func (m *migration) Extension() string {
	return m.ext
}

func (m *migration) IsDev() bool {
	return m.dev
}

func (m *migration) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return m.db.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS migrations (
			name VARCHAR(100) NOT NULL,
			stage VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(name, stage)
		);`,
	)
}

func (m *migration) Summary() (Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := m.db.Scan(ctx, `SELECT name, stage, created_at FROM migrations ORDER BY created_at ASC;`)
	if err != nil {
		return nil, err
	}

	result := make(Summary, 0)
	defer rows.Close()
	for rows.Next() {
		var name, stage string
		var createdAt time.Time
		err := rows.Scan(&name, &stage, &createdAt)
		if err != nil {
			return nil, err
		}
		result = append(result, Migrated{
			Name:      name,
			Stage:     stage,
			CreatedAt: createdAt,
		})
	}
	return result, nil
}

func (m *migration) Up(stages []string, options ...MigrationOption) (Summary, error) {
	if len(stages) == 0 {
		return nil, nil
	}

	// Hot reload on dev mode
	if m.dev {
		if err := m.Load(); err != nil {
			return nil, err
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create option
	option := newOption()
	for _, opt := range options {
		opt(option)
	}

	// Read migrated files
	migrated, err := m.Summary()
	if err != nil {
		return nil, err
	}

	// Filter files
	files := m.files.Filter(option.only.Elements(), option.exclude.Elements())
	if files.Len() == 0 {
		return nil, nil
	}

	// Execute scripts
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	result := make(Summary, 0)
	err = m.db.Transaction(ctx, func(tx ExecutableScanner) error {
		for _, stage := range stages {
			for _, file := range files {
				if migrated.includes(file.name, stage) {
					continue
				}

				script, ok := file.UpScript(stage)
				if !ok || len(script) == 0 {
					continue
				}

				err := tx.Exec(ctx, script)
				if err != nil {
					return fmt.Errorf("%s: %w", file.name, err)
				}

				err = tx.Exec(
					ctx,
					fmt.Sprintf(`INSERT INTO migrations (name, stage) VALUES ('%s', '%s');`, file.name, stage),
				)
				if err != nil {
					return fmt.Errorf("%s: %w", file.name, err)
				}

				result = append(result, Migrated{
					Stage:     stage,
					Name:      file.name,
					CreatedAt: time.Now(),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *migration) Down(stages []string, options ...MigrationOption) (Summary, error) {
	if len(stages) == 0 {
		return nil, nil
	}

	// Hot reload on dev mode
	if m.dev {
		if err := m.Load(); err != nil {
			return nil, err
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create option
	option := newOption()
	for _, opt := range options {
		opt(option)
	}

	// Read migrated files
	migrated, err := m.Summary()
	if err != nil {
		return nil, err
	}

	// Filter files
	files := m.files.Reverse().Filter(option.only.Elements(), option.exclude.Elements())
	if files.Len() == 0 {
		return nil, nil
	}

	// Execute scripts
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	result := make(Summary, 0)
	err = m.db.Transaction(ctx, func(tx ExecutableScanner) error {
		for _, stage := range stages {
			for _, file := range files {
				if !migrated.includes(file.name, stage) {
					continue
				}

				script, ok := file.DownScript(stage)
				if !ok {
					continue
				}

				if len(script) != 0 {
					if err := tx.Exec(ctx, script); err != nil {
						return fmt.Errorf("%s: %w", file.name, err)
					}
				}

				err = tx.Exec(
					ctx,
					fmt.Sprintf(`DELETE FROM migrations WHERE name = '%s' AND stage = '%s';`, file.name, stage),
				)
				if err != nil {
					return fmt.Errorf("%s: %w", file.name, err)
				}

				result = append(result, Migrated{
					Stage:     stage,
					Name:      file.name,
					CreatedAt: time.Now(),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *migration) Refresh(stages []string, options ...MigrationOption) (Summary, error) {
	if len(stages) == 0 {
		return nil, nil
	}

	// Hot reload on dev mode
	if m.dev {
		if err := m.Load(); err != nil {
			return nil, err
		}
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create option
	option := newOption()
	for _, opt := range options {
		opt(option)
	}

	// Read migrated files
	migrated, err := m.Summary()
	if err != nil {
		return nil, err
	}

	// Filter upFiles
	downFiles := m.files.Reverse().Filter(option.only.Elements(), option.exclude.Elements())
	upFiles := m.files.Filter(option.only.Elements(), option.exclude.Elements())

	// Execute scripts
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	result := make(Summary, 0)
	err = m.db.Transaction(ctx, func(tx ExecutableScanner) error {
		for _, stage := range stages {
			// Down
			for _, file := range downFiles {
				if !migrated.includes(file.name, stage) {
					continue
				}

				script, ok := file.DownScript(stage)
				if !ok {
					continue
				}

				if len(script) != 0 {
					if err := tx.Exec(ctx, script); err != nil {
						return fmt.Errorf(`rollback "%s": %w`, file.name, err)
					}
				}

				err = tx.Exec(
					ctx,
					fmt.Sprintf(`DELETE FROM migrations WHERE name = '%s' AND stage = '%s';`, file.name, stage),
				)
				if err != nil {
					return fmt.Errorf(`rollback "%s": %w`, file.name, err)
				}
			}

			// Up
			for _, file := range upFiles {
				script, ok := file.UpScript(stage)
				if !ok || len(script) == 0 {
					continue
				}

				err := tx.Exec(ctx, script)
				if err != nil {
					return fmt.Errorf(`up "%s": %w`, file.name, err)
				}

				err = tx.Exec(
					ctx,
					fmt.Sprintf(`INSERT INTO migrations (name, stage) VALUES ('%s', '%s');`, file.name, stage),
				)
				if err != nil {
					return fmt.Errorf(`up "%s": %w`, file.name, err)
				}

				result = append(result, Migrated{
					Stage:     stage,
					Name:      file.name,
					CreatedAt: time.Now(),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
