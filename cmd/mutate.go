package cmd

import (
	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/client"
	"github.com/looplj/graphql-cli/internal/config"
	"github.com/looplj/graphql-cli/internal/printer"
)

var mutateCmd = &cobra.Command{
	Use:   "mutate <graphql-mutation>",
	Short: "Execute a GraphQL mutation",
	Long: `Execute a GraphQL mutation against the configured endpoint.

Examples:
  graphql-cli mutate 'mutation { createUser(name: "test") { id } }'
  graphql-cli mutate -f mutation.graphql -v '{"name": "test"}'
  graphql-cli mutate -f mutation.graphql -H "Authorization=Bearer token"`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: requireEndpoint,
	RunE:    runMutate,
}

var (
	mutateFile      string
	mutateVariables string
	mutateHeaders   []string
)

func init() {
	mutateCmd.Flags().StringVarP(&mutateFile, "file", "f", "", "read mutation from file")
	mutateCmd.Flags().StringVarP(&mutateVariables, "variables", "v", "", "mutation variables as JSON string")
	mutateCmd.Flags().StringSliceVarP(&mutateHeaders, "header", "H", nil, "extra HTTP headers (key=value), can be specified multiple times")
	rootCmd.AddCommand(mutateCmd)
}

func runMutate(cmd *cobra.Command, args []string) error {
	q, err := resolveGraphQL(args, mutateFile)
	if err != nil {
		return err
	}

	vars, err := parseVariables(mutateVariables)
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	ep, err := cfg.GetEndpoint(endpointName)
	if err != nil {
		return err
	}

	headers, err := parseHeaders(mutateHeaders)
	if err != nil {
		return err
	}

	resp, err := client.Execute(ep, q, vars, headers)
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		printer.PrintErrors(resp.Errors)
	}

	if resp.Data != nil {
		printer.PrintJSON(resp.Data)
	}

	return nil
}
