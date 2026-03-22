package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout [endpoint]",
	Short: "Remove stored credentials for an endpoint",
	Long: `Remove stored credentials for a GraphQL endpoint.

Examples:
  graphql-cli logout              # logout from default endpoint
  graphql-cli logout production   # logout from specific endpoint`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	epName, err := resolveEndpointName(args)
	if err != nil {
		return err
	}

	ep, err := cfg.GetEndpoint(epName)
	if err != nil {
		return err
	}

	store := auth.NewStore()
	if err := store.Delete(ep.Name); err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}

	successColor := color.New(color.FgGreen, color.Bold)
	successColor.Printf("✓ ")
	fmt.Printf("Logged out from %q\n", ep.Name)

	return nil
}
