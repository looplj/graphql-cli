package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Client is a reusable GraphQL client that caches parsed schemas in memory.
type Client struct {
	mu         sync.RWMutex
	schemas    map[string]*Schema
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) {
		cl.httpClient = c
	}
}

// New creates a new Client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		schemas:    make(map[string]*Schema),
		httpClient: http.DefaultClient,
	}
	for _, o := range opts {
		o(c)
	}

	return c
}

// Request represents a GraphQL request payload.
type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// Response represents a GraphQL response.
//
//nolint:errname // Checked.
type Response struct {
	Data   json.RawMessage `json:"data"`
	Errors []ResponseError `json:"errors"`
}

// ResponseError represents a GraphQL error in the response.
type ResponseError struct {
	Message string `json:"message"`
}

// HasErrors returns true if the response contains any GraphQL errors.
func (r *Response) HasErrors() bool {
	return len(r.Errors) > 0
}

// Error returns the first GraphQL error message, or empty string if none.
func (r *Response) Error() string {
	if len(r.Errors) == 0 {
		return ""
	}

	msgs := make([]string, len(r.Errors))
	for i, e := range r.Errors {
		msgs[i] = e.Message
	}

	return strings.Join(msgs, "; ")
}

// Execute sends a GraphQL request to the given URL.
func (c *Client) Execute(ctx context.Context, url string, req Request, headers map[string]string) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed Response
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &parsed, nil
}
