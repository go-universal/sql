package query

import "strings"

type Options func(*queryManager)

// WithRoot sets the root directory for SQL query files in the filesystem.
func WithRoot(root string) Options {
	root = normalizePath(root)
	return func(q *queryManager) {
		if root != "" {
			q.root = root + "/"
		} else {
			q.root = "."
		}
	}
}

// WithExtension sets the file extension for SQL query files.
func WithExtension(ext string) Options {
	ext = strings.TrimSpace(ext)
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return func(q *queryManager) {
		if ext != "" {
			q.ext = ext
		}
	}
}

// WithEnv sets the environment mode. In development mode,
// Load() is called every time a query is retrieved.
// Avoid enabling this in production for performance reasons.
func WithEnv(isDev bool) Options {
	return func(q *queryManager) {
		q.dev = isDev
	}
}

// WithResolver assigns a custom resolver for handling placeholders in SQL queries.
func WithResolver(resolver PlaceholderResolver) Options {
	return func(q *queryManager) {
		q.resolver = resolver
	}
}
