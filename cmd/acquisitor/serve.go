package acquisitor

import (
	"log"

	"github.com/prutonis/acquisitor/internal/srv"
	"github.com/spf13/cobra"
)

func init() {
	// initConfig reads in config file and ENV variables if set.
	initConfig()
	log.Printf("Serve Init called")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the acquisitor service",
	Long:  `Starts the acquisitor service`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.StartServer()
	},
}
