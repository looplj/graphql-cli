package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	endpointName string
)

var rootCmd = &cobra.Command{
	Use:   "graphql-cli",
	Short: "A CLI tool for exploring and querying GraphQL APIs",
	Long: `graphql-cli supports configuring multiple GraphQL endpoints (remote URL or local schema file),
and provides subcommands to explore schemas and execute queries/mutations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/graphql-cli/config.yaml)")
}

func addEndpointFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&endpointName, "endpoint", "e", "", "endpoint name to use")
}

// requireEndpoint is a PreRunE that ensures -e is specified.
var requireEndpoint = func(cmd *cobra.Command, args []string) error {
	if endpointName == "" {
		return fmt.Errorf("endpoint is required, use -e to specify one (see 'graphql-cli endpoint list' for available endpoints)")
	}

	return nil
}
