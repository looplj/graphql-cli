package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// Schema wraps a parsed GraphQL schema with its SDL source.
type Schema struct {
	SDL    string
	Parsed *ast.Schema
}

// LoadSchemaFromFile loads and parses a GraphQL schema from a file, caching the result.
// Subsequent calls with the same key return the cached schema.
func (c *Client) LoadSchemaFromFile(key, path string) (*Schema, error) {
	if s := c.getCachedSchema(key); s != nil {
		return s, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema file: %w", err)
	}

	return c.parseAndCache(key, string(data))
}

// LoadSchemaFromURL loads a schema via introspection from a URL, caching the result.
// Subsequent calls with the same key return the cached schema.
func (c *Client) LoadSchemaFromURL(ctx context.Context, key, url string, headers map[string]string) (*Schema, error) {
	if s := c.getCachedSchema(key); s != nil {
		return s, nil
	}

	sdl, err := c.introspect(ctx, url, headers)
	if err != nil {
		return nil, err
	}

	return c.parseAndCache(key, sdl)
}

// LoadSchemaFromSDL parses the given SDL string and caches it under the given key.
// Subsequent calls with the same key return the cached schema.
func (c *Client) LoadSchemaFromSDL(key, sdl string) (*Schema, error) {
	if s := c.getCachedSchema(key); s != nil {
		return s, nil
	}

	return c.parseAndCache(key, sdl)
}

// InvalidateSchema removes a cached schema by key.
func (c *Client) InvalidateSchema(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.schemas, key)
}

// InvalidateAllSchemas removes all cached schemas.
func (c *Client) InvalidateAllSchemas() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.schemas = make(map[string]*Schema)
}

func (c *Client) getCachedSchema(key string) *Schema {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if s, ok := c.schemas[key]; ok {
		return s
	}

	return nil
}

func (c *Client) parseAndCache(key, sdl string) (*Schema, error) {
	parsed, err := gqlparser.LoadSchema(&ast.Source{
		Name:  key,
		Input: sdl,
	})
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	s := &Schema{SDL: sdl, Parsed: parsed}

	c.mu.Lock()
	c.schemas[key] = s
	c.mu.Unlock()

	return s, nil
}

const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args { ...InputValue }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args { ...InputValue }
    type { ...TypeRef }
    isDeprecated
    deprecationReason
  }
  inputFields { ...InputValue }
  interfaces { ...TypeRef }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes { ...TypeRef }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
            }
          }
        }
      }
    }
  }
}
`

func (c *Client) introspect(ctx context.Context, url string, headers map[string]string) (string, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("introspection request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("introspection failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Schema json.RawMessage `json:"__schema"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse introspection response: %w", err)
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("introspection error: %s", result.Errors[0].Message)
	}

	return introspectionToSDL(result.Data.Schema)
}
