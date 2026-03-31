package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/config"
)

var (
	updateURL         string
	updateDescription string
	updateHeaders     []string
)

func init() {
	endpointCmd.AddCommand(newUpdateCmd())
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an existing endpoint's URL or headers",
		Long: `Update an existing GraphQL endpoint configuration.

Examples:
  graphql-cli endpoint update production --url https://api.example.com/v2/graphql
  graphql-cli endpoint update production --header "Authorization=Bearer new-token"
  graphql-cli endpoint update production --url https://new-url.com/graphql --header "X-Custom=value"`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().StringVar(&updateURL, "url", "", "new GraphQL endpoint URL")
	cmd.Flags().StringVarP(&updateDescription, "description", "d", "", "new endpoint description")
	cmd.Flags().StringSliceVar(&updateHeaders, "header", nil, "HTTP headers to add/update (key=value), can be specified multiple times")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

	var urlPtr *string

	if cmd.Flags().Changed("url") {
		u, err := url.Parse(updateURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("invalid endpoint URL %q: must be a valid http or https URL", updateURL)
		}

		urlPtr = &updateURL
	}

	var descPtr *string
	if cmd.Flags().Changed("description") {
		descPtr = &updateDescription
	}

	headers := make(map[string]string)

	for _, h := range updateHeaders {
		k, v, ok := parseHeader(h)
		if !ok {
			return fmt.Errorf("invalid header format %q, expected key=value", h)
		}

		headers[k] = v
	}

	if urlPtr == nil && descPtr == nil && len(headers) == 0 {
		return fmt.Errorf("must specify at least one of --url, --description, or --header")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if err := cfg.UpdateEndpoint(name, urlPtr, descPtr, headers); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return err
	}

	fmt.Printf("Updated endpoint %q\n", name)

	return nil
}
