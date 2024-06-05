/*
Copyright © 2024 IonM
*/
package acquisitor

import (
	"fmt"
	"os"

	"runtime/debug"

	"github.com/pkg/errors"
	cfg "github.com/prutonis/acquisitor/internal/cfg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var config = &cfg.AcqConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "acquisitor",
	Short: "Data acquisition tool. It collects data from different sensors and sends it to a server for storage and analysis",
	Long: `Acquisitor runs as service and collects data 
from different sensors and sends it to a server 
for storage and analysis. 
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.acquisitor.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".acquisitor" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".acquisitor")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if err := viper.Unmarshal(config); err != nil {
		fmt.Println("Unable to decode into struct:", errors.WithStack(err))
		debug.PrintStack()
	}
}
