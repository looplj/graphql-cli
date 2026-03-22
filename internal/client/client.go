package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/looplj/graphql-cli/internal/auth"
	"github.com/looplj/graphql-cli/internal/config"
	"github.com/looplj/graphql-cli/internal/printer"
)

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type Response struct {
	Data   json.RawMessage        `json:"data"`
	Errors []printer.GraphQLError `json:"errors"`
}

// Execute sends a GraphQL request.
// Header priority: config headers < stored credential < extraHeaders (CLI -H flags).
func Execute(ep *config.Endpoint, query string, variables map[string]any, extraHeaders map[string]string) (*Response, error) {
	if ep.URL == "" {
		return nil, fmt.Errorf("endpoint %q has no URL configured; only local schema_file is available (use 'find' to explore the schema)", ep.Name)
	}

	reqBody, err := json.Marshal(Request{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ep.URL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 1. config file headers (lowest priority)
	for k, v := range ep.Headers {
		req.Header.Set(k, v)
	}

	// 2. stored credential headers (middle priority)
	store := auth.NewStore()
	if cred, _ := store.Load(ep.Name); cred != nil {
		for k, v := range cred.AuthHeaders() {
			req.Header.Set(k, v)
		}
	}

	// 3. CLI -H flag headers (highest priority)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}
