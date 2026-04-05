package schema

import (
	"context"
	"fmt"
	"maps"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
	"github.com/looplj/graphql-cli/pkg/graphql"
)

// LoadSDL loads a schema SDL string from a config endpoint.
// It delegates to the public graphql.Client for actual loading.
func LoadSDL(ep *config.Endpoint) (string, error) {
	c := graphql.New()

	if ep.SchemaFile != "" {
		s, err := c.LoadSchemaFromFile(ep.Name, ep.SchemaFile)
		if err != nil {
			return "", err
		}

		return s.SDL, nil
	}

	if ep.URL != "" {
		headers := buildHeaders(ep)

		s, err := c.LoadSchemaFromURL(context.Background(), ep.Name, ep.URL, headers)
		if err != nil {
			return "", err
		}

		return s.SDL, nil
	}

	return "", fmt.Errorf("endpoint %q has neither url nor schema_file", ep.Name)
}

func buildHeaders(ep *config.Endpoint) map[string]string {
	headers := make(map[string]string)

	maps.Copy(headers, ep.Headers)

	store := auth.NewStore()
	if cred, _ := store.Load(ep.Name); cred != nil {
		maps.Copy(headers, cred.AuthHeaders())
	}

	return headers
}
