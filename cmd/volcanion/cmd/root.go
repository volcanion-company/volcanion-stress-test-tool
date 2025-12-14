package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	apiBaseURL string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "volcanion",
	Short: "Volcanion Stress Test Tool CLI",
	Long: `Volcanion is a powerful HTTP load testing tool.
	
Use this CLI to create, manage, and run stress tests from the command line.
Perfect for CI/CD integration and automated performance testing.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.volcanion.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiBaseURL, "api", "http://localhost:8080", "API base URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			return
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".volcanion")
	}

	viper.SetEnvPrefix("VOLCANION")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetAPIBaseURL returns the configured API base URL
func GetAPIBaseURL() string {
	return viper.GetString("api")
}

// IsVerbose returns whether verbose output is enabled
func IsVerbose() bool {
	return viper.GetBool("verbose")
}
