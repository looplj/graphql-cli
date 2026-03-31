package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
)

var listDetail bool

func init() {
	endpointCmd.AddCommand(newListCmd())
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured endpoints",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}

	cmd.Flags().BoolVar(&listDetail, "detail", false, "show endpoint details including headers and auth")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if len(cfg.Endpoints) == 0 {
		fmt.Println("No endpoints configured. Use 'graphql-cli endpoint add' to add one.")
		return nil
	}

	nameColor := color.New(color.FgGreen, color.Bold)
	dimColor := color.New(color.FgHiBlack)
	authColor := color.New(color.FgCyan)

	store := auth.NewStore()

	for _, ep := range cfg.Endpoints {
		nameColor.Println(ep.Name)

		if ep.URL != "" {
			dimColor.Printf("  URL: %s\n", ep.URL)
		}

		if ep.Description != "" {
			fmt.Printf("  %s\n", ep.Description)
		}

		if listDetail {
			if ep.SchemaFile != "" {
				dimColor.Printf("  Schema: %s\n", ep.SchemaFile)
			}

			if len(ep.Headers) > 0 {
				dimColor.Println("  Headers:")

				for k, v := range ep.Headers {
					dimColor.Printf("    %s: %s\n", k, maskValue(v))
				}
			}

			if cred, _ := store.Load(ep.Name); cred != nil {
				authColor.Printf("  Auth: %s\n", cred.String())
			}
		}
	}

	return nil
}

func maskValue(v string) string {
	if len(v) <= 8 {
		return strings.Repeat("*", len(v))
	}

	return v[:4] + strings.Repeat("*", len(v)-4)
}
