package migration

type migrationOption struct {
	only    *optionSet
	exclude *optionSet
}

func newOption() *migrationOption {
	return &migrationOption{
		only:    &optionSet{elements: make([]string, 0)},
		exclude: &optionSet{elements: make([]string, 0)},
	}
}

type MigrationOption func(*migrationOption)

// OnlyFiles specifies the files to include in the migration.
func OnlyFiles(files ...string) MigrationOption {
	return func(o *migrationOption) {
		o.only.Add(files...)
	}
}

// SkipFiles specifies the files to exclude from the migration.
func SkipFiles(files ...string) MigrationOption {
	return func(o *migrationOption) {
		o.exclude.Add(files...)
	}
}
