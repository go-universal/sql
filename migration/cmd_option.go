package migration

type cliOption struct {
	root      string
	create    bool
	stages    *optionSet
	refreshes *optionSet
	only      *optionSet
	exclude   *optionSet
	callback  func()
}

func newCLIOption() *cliOption {
	return &cliOption{
		root:      "",
		create:    false,
		stages:    &optionSet{elements: make([]string, 0)},
		refreshes: &optionSet{elements: make([]string, 0)},
		only:      &optionSet{elements: make([]string, 0)},
		exclude:   &optionSet{elements: make([]string, 0)},
	}
}

type CLIOptions func(*cliOption)

// WithOutputPath sets the output file root path for the new command.
func WithOutputPath(path string) CLIOptions {
	path = normalizePath(path)
	return func(o *cliOption) {
		o.root = path
	}
}

// WithDefaultStages adds stages to auto-run, generate, and rollback on CLI.
func WithDefaultStages(stages ...string) CLIOptions {
	return func(o *cliOption) {
		o.stages.Add(stages...)
	}
}

// WithRefreshStages adds the stages to run on refresh.
func WithRefreshStages(stages ...string) CLIOptions {
	return func(o *cliOption) {
		o.refreshes.Add(stages...)
	}
}

// WithOnlyFiles specifies the files to include in the migration.
func WithOnlyFiles(files ...string) CLIOptions {
	return func(o *cliOption) {
		o.only.Add(files...)
	}
}

// WithSkipFiles specifies the files to exclude from the migration.
func WithSkipFiles(files ...string) CLIOptions {
	return func(o *cliOption) {
		o.exclude.Add(files...)
	}
}

// WithNewCMD enables a new command to create a migration file in development mode.
func WithNewCMD(enabled bool) CLIOptions {
	return func(o *cliOption) {
		o.create = enabled
	}
}

// WithCallback register a callback function to call after command finished.
func WithCallback(cb func()) CLIOptions {
	return func(o *cliOption) {
		o.callback = cb
	}
}
