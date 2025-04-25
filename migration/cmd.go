package migration

import "github.com/spf13/cobra"

// NewMigrationCLI creates a new cobra command for database migration with the provided options.
func NewMigrationCLI(m Migration, options ...CLIOptions) *cobra.Command {
	option := newCLIOption()
	for _, opt := range options {
		opt(option)
	}

	cmd := &cobra.Command{
		Use:   "migration",
		Short: "migrate database",
	}
	if option.create {
		cmd.AddCommand(cmdNew(m, option))
	}
	cmd.AddCommand(cmdUp(m, option))
	cmd.AddCommand(cmdDown(m, option))
	cmd.AddCommand(cmdRefresh(m, option))
	cmd.AddCommand(cmdSummary(m, option))
	return cmd
}
