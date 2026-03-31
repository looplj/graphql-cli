package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/config"
)

var (
	addURL         string
	addSchemaFile  string
	addDescription string
	addHeaders     []string
)

func init() {
	endpointCmd.AddCommand(newAddCmd())
}

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new endpoint to the configuration",
		Long: `Add a new GraphQL endpoint. Specify either a remote URL or a local schema file.

Examples:
  graphql-cli endpoint add production --url https://api.example.com/graphql --header "Authorization=Bearer token"
  graphql-cli endpoint add local --schema-file ./schema.graphql --description "Local dev schema"`,
		Args: cobra.ExactArgs(1),
		RunE: runAdd,
	}

	cmd.Flags().StringVar(&addURL, "url", "", "GraphQL endpoint URL")
	cmd.Flags().StringVar(&addSchemaFile, "schema-file", "", "path to local GraphQL schema file")
	cmd.Flags().StringVarP(&addDescription, "description", "d", "", "endpoint description")
	cmd.Flags().StringSliceVar(&addHeaders, "header", nil, "HTTP headers (key=value), can be specified multiple times")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	if addURL == "" && addSchemaFile == "" {
		return fmt.Errorf("must specify either --url or --schema-file")
	}

	if addURL != "" {
		u, err := url.Parse(addURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("invalid endpoint URL %q: must be a valid http or https URL", addURL)
		}
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	headers := make(map[string]string)

	for _, h := range addHeaders {
		k, v, ok := parseHeader(h)
		if !ok {
			return fmt.Errorf("invalid header format %q, expected key=value", h)
		}

		headers[k] = v
	}

	ep := config.Endpoint{
		Name:        name,
		Description: addDescription,
		URL:         addURL,
		SchemaFile:  addSchemaFile,
		Headers:     headers,
	}

	if err := cfg.AddEndpoint(ep); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return err
	}

	fmt.Printf("Added endpoint %q\n", name)

	return nil
}

func parseHeader(s string) (string, string, bool) {
	for i, c := range s {
		if c == '=' {
			return s[:i], s[i+1:], true
		}
	}

	return "", "", false
}
