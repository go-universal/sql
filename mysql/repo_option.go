package mysql

type options struct {
	only    []string
	exclude []string
}

type RepositoryOption func(*options)

// OnlyFields returns a RepositoryOption function that specifies which fields to include.
func OnlyFields(fields ...string) RepositoryOption {
	return func(o *options) {
		o.only = append(o.only, fields...)
	}
}

// SkipFields returns a RepositoryOption function that specifies which fields to exclude.
func SkipFields(fields ...string) RepositoryOption {
	return func(o *options) {
		o.exclude = append(o.exclude, fields...)
	}
}
