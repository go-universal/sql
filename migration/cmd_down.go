package migration

import (
	"fmt"
	"strings"

	"github.com/go-universal/console"
	"github.com/spf13/cobra"
)

func cmdDown(m Migration, option *cliOption) *cobra.Command {
	downCmd := &cobra.Command{}
	downCmd.Use = "down [stage1, stage2, ...]"
	downCmd.Short = "rollback migrations"
	downCmd.Flags().StringP("name", "n", "", "migration name")
	downCmd.Run = func(cmd *cobra.Command, args []string) {
		if option.callback != nil {
			defer option.callback()
		}

		stages := append([]string{}, args...)
		if len(stages) == 0 {
			stages = option.stages.Elements()
		}

		if len(stages) == 0 {
			console.Message().
				Red("Down").Italic().
				Print("no stage stage specified")
			return
		}

		options := make([]MigrationOption, 0)
		if name := getFlag(cmd, "name"); name != "" {
			options = append(options, OnlyFiles(name))
		}

		result, err := m.Down(stages, options...)
		if err != nil {
			console.Message().Red("Down").Italic().Print(err.Error())
			return
		}

		console.PrintF("@Bwb{ Rollback Summery: }\n")
		if result.IsEmpty() {
			console.Message().Indent().Italic().Print("nothing to roll back")
		} else {
			for stage, files := range result.GroupByStage() {
				console.PrintF("@BUb{%s} @b{Stage} @Ib{(%d Files)}:\n", strings.ToTitle(stage), len(files))
				for _, file := range files {
					console.PrintF("    @g{DOWN:} @I{%s}\n", file.Name)
				}

				fmt.Println()
			}
		}
	}

	return downCmd
}
