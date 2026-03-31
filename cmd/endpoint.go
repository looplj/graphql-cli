package cmd

import "github.com/spf13/cobra"

var endpointCmd = &cobra.Command{
	Use:     "endpoint",
	Aliases: []string{"endpoints"},
	Short:   "Manage configured GraphQL endpoints",
}

func init() {
	rootCmd.AddCommand(endpointCmd)
}
