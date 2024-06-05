package acquisitor

import (
	"github.com/prutonis/acquisitor/internal/insp"
	"github.com/spf13/cobra"
)

func init() {
	// initConfig reads in config file and ENV variables if set.
	initConfig()
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspects the acquisitor service",
	Long:  `Inspects the acquisitor service`,
	Run: func(cmd *cobra.Command, args []string) {
		insp.Inspect()
	},
}
