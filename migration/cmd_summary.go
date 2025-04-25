package migration

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-universal/console"
	"github.com/spf13/cobra"
)

func cmdSummary(m Migration, option *cliOption) *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "show migration summary",
		Run: func(cmd *cobra.Command, args []string) {
			if option.callback != nil {
				defer option.callback()
			}

			summary, err := m.Summary()
			if err != nil {
				console.Message().Red("Summary").Italic().Print(err.Error())
				return
			}

			if summary.IsEmpty() {
				console.Message().Blue("Summary").Italic().Print("nothing migrated!")
				return
			}

			console.PrintF("@Bwb{ Migration Summery: }\n")
			for stage, files := range summary.GroupByStage() {
				console.PrintF("@BUb{%s} @b{Stage} @Ib{(%d Files)}:\n", strings.ToTitle(stage), len(files))
				for _, file := range files {
					console.PrintF("    @g{%s}: @I{%s}\n", file.Name, humanize.Time(file.CreatedAt))
				}

				fmt.Println()
			}
		},
	}
}
