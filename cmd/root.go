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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/graphql-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&endpointName, "endpoint", "e", "", "endpoint name to use")
}

// requireEndpoint is a PreRunE that ensures -e is specified.
var requireEndpoint = func(cmd *cobra.Command, args []string) error {
	if endpointName == "" {
		return fmt.Errorf("endpoint is required, use -e to specify one (see 'graphql-cli list' for available endpoints)")
	}

	return nil
}

// resolveEndpointName returns the endpoint name from positional arg or -e flag.
func resolveEndpointName(args []string) (string, error) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		name = endpointName
	}

	if name == "" {
		return "", fmt.Errorf("endpoint is required, specify as argument or use -e (see 'graphql-cli list' for available endpoints)")
	}

	return name, nil
}
