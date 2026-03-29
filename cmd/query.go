package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/client"
	"github.com/looplj/graphql-cli/internal/config"
	"github.com/looplj/graphql-cli/internal/printer"
)

var queryCmd = &cobra.Command{
	Use:   "query <graphql-query>",
	Short: "Execute a GraphQL query",
	Long: `Execute a GraphQL query against the configured endpoint.

Examples:
  graphql-cli query '{ users { id name } }'
  graphql-cli query -f query.graphql
  graphql-cli query '{ user(id: "1") { name } }' -v '{"id": "1"}'
  graphql-cli query '{ me { name } }' -H "Authorization=Bearer token"`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: requireEndpoint,
	RunE:    runQuery,
}

var (
	queryFile      string
	queryVariables string
	queryHeaders   []string
)

func init() {
	addEndpointFlag(queryCmd)
	queryCmd.Flags().StringVarP(&queryFile, "file", "f", "", "read query from file")
	queryCmd.Flags().StringVarP(&queryVariables, "variables", "v", "", "query variables as JSON string")
	queryCmd.Flags().StringSliceVarP(&queryHeaders, "header", "H", nil, "extra HTTP headers (key=value), can be specified multiple times")
	rootCmd.AddCommand(queryCmd)
}

func runQuery(cmd *cobra.Command, args []string) error {
	q, err := resolveGraphQL(args, queryFile)
	if err != nil {
		return err
	}

	vars, err := parseVariables(queryVariables)
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

	headers, err := parseHeaders(queryHeaders)
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

func resolveGraphQL(args []string, file string) (string, error) {
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}

		return string(data), nil
	}

	if len(args) > 0 {
		return args[0], nil
	}

	return "", fmt.Errorf("provide a query string or use -f to read from file")
}

func parseHeaders(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	headers := make(map[string]string, len(raw))
	for _, h := range raw {
		k, v, ok := parseHeader(h)
		if !ok {
			return nil, fmt.Errorf("invalid header format %q, expected key=value", h)
		}

		headers[k] = v
	}

	return headers, nil
}

func parseVariables(s string) (map[string]any, error) {
	if s == "" {
		return nil, nil
	}

	var vars map[string]any
	if err := json.Unmarshal([]byte(s), &vars); err != nil {
		return nil, fmt.Errorf("parse variables: %w", err)
	}

	return vars, nil
}
