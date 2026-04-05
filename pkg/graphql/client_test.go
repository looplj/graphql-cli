package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	c := New()

	resp, err := c.Execute(context.Background(), server.URL, Request{
		Query: "{ ok }",
	}, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if string(resp.Data) != `{"ok":true}` {
		t.Fatalf("unexpected data: %s", resp.Data)
	}
}

func TestExecuteWithHeaders(t *testing.T) {
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":null}`))
	}))
	defer server.Close()

	c := New()

	_, err := c.Execute(context.Background(), server.URL, Request{Query: "{ me }"}, map[string]string{
		"Authorization": "Bearer test-token",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected Authorization header, got %q", gotAuth)
	}
}

func TestLoadSchemaFromFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/schema.graphql"

	sdl := "type Query { hello: String }"
	if err := os.WriteFile(path, []byte(sdl), 0o644); err != nil {
		t.Fatal(err)
	}

	c := New()

	s1, err := c.LoadSchemaFromFile("test", path)
	if err != nil {
		t.Fatalf("LoadSchemaFromFile returned error: %v", err)
	}

	if s1.SDL != sdl {
		t.Fatalf("unexpected SDL: %s", s1.SDL)
	}

	if s1.Parsed == nil {
		t.Fatal("expected parsed schema to be non-nil")
	}

	// second call should return cached
	s2, err := c.LoadSchemaFromFile("test", path)
	if err != nil {
		t.Fatalf("second LoadSchemaFromFile returned error: %v", err)
	}

	if s2.Parsed != s1.Parsed {
		t.Fatal("expected cached schema to be returned")
	}
}

func TestLoadSchemaFromSDL(t *testing.T) {
	c := New()

	sdl := "type Query { hello: String }"

	s, err := c.LoadSchemaFromSDL("test", sdl)
	if err != nil {
		t.Fatalf("LoadSchemaFromSDL returned error: %v", err)
	}

	if s.Parsed.Query == nil {
		t.Fatal("expected Query type to be parsed")
	}
}

func TestInvalidateSchema(t *testing.T) {
	c := New()

	sdl := "type Query { hello: String }"

	_, err := c.LoadSchemaFromSDL("test", sdl)
	if err != nil {
		t.Fatal(err)
	}

	c.InvalidateSchema("test")

	// should re-parse after invalidation
	s, err := c.LoadSchemaFromSDL("test", sdl)
	if err != nil {
		t.Fatal(err)
	}

	if s.Parsed == nil {
		t.Fatal("expected schema to be re-parsed")
	}
}

func TestFind(t *testing.T) {
	c := New()

	sdl := `
type Query {
  users: [User!]!
  posts: [Post!]!
}
type User {
  id: ID!
  name: String!
}
type Post {
  id: ID!
  title: String!
}
enum UserRole {
  ADMIN
  USER
}
`

	s, err := c.LoadSchemaFromSDL("test", sdl)
	if err != nil {
		t.Fatalf("LoadSchemaFromSDL returned error: %v", err)
	}

	results, err := Find(s, "user", FindScope{})
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	found := false

	for _, r := range results {
		if r.Name == "User" && r.Kind == "type" {
			found = true
		}
	}

	if !found {
		t.Fatal("expected to find User type")
	}
}

func TestLoadSchemaFromURL(t *testing.T) {
	introspectionResp := map[string]any{
		"data": map[string]any{
			"__schema": map[string]any{
				"queryType":        map[string]string{"name": "Query"},
				"mutationType":     nil,
				"subscriptionType": nil,
				"types": []map[string]any{
					{
						"kind":        "OBJECT",
						"name":        "Query",
						"description": "",
						"fields": []map[string]any{
							{
								"name":              "hello",
								"description":       "",
								"args":              []any{},
								"type":              map[string]any{"kind": "SCALAR", "name": "String", "ofType": nil},
								"isDeprecated":      false,
								"deprecationReason": nil,
							},
						},
						"inputFields":   nil,
						"interfaces":    []any{},
						"enumValues":    nil,
						"possibleTypes": nil,
					},
				},
				"directives": []any{},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(introspectionResp)
	}))
	defer server.Close()

	c := New()

	s, err := c.LoadSchemaFromURL(context.Background(), "test-url", server.URL, nil)
	if err != nil {
		t.Fatalf("LoadSchemaFromURL returned error: %v", err)
	}

	if s.SDL == "" {
		t.Fatal("expected non-empty SDL")
	}
}
