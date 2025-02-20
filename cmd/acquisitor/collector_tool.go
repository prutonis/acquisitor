package acquisitor

import (
	"github.com/prutonis/acquisitor/internal/clct"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(collectCmd)
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect data and send it to the server",
	Long:  `Collect data and send it to the server`,
	Run: func(cmd *cobra.Command, args []string) {
		clct.Collect()
	},
}
