package migration

import (
	"strings"
)

type Option func(*migration)

// WithRoot sets the root directory for migration files.
func WithRoot(root string) Option {
	root = normalizePath(root)
	return func(q *migration) {
		if root != "" {
			q.root = root
		} else {
			q.root = "."
		}
	}
}

// WithExtension sets the file extension for migration files.
func WithExtension(ext string) Option {
	ext = strings.TrimSpace(ext)
	ext = strings.TrimLeft(ext, ".")
	return func(q *migration) {
		if ext != "" {
			q.ext = ext
		}
	}
}

// WithEnv sets the environment mode for migrations.
// Enables development mode if true, causing Load() to be called on each migration run.
func WithEnv(isDev bool) Option {
	return func(q *migration) {
		q.dev = isDev
	}
}
