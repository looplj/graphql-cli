package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/looplj/graphql-cli/internal/audit"
	"github.com/looplj/graphql-cli/internal/config"
)

func TestExecuteWritesAuditLog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	ep := &config.Endpoint{Name: "production", URL: server.URL}
	statement := "query Viewer { viewer { id } }"

	resp, err := Execute(ep, statement, nil, nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if string(resp.Data) != `{"ok":true}` {
		t.Fatalf("unexpected response data: %s", resp.Data)
	}

	data, err := os.ReadFile(config.DefaultAuditLogPath())
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 audit log line, got %d", len(lines))
	}

	var entry audit.Entry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("parse audit log entry: %v", err)
	}

	if entry.Endpoint != ep.Name {
		t.Fatalf("expected endpoint %q, got %q", ep.Name, entry.Endpoint)
	}

	if entry.Status != "success" {
		t.Fatalf("expected success status, got %q", entry.Status)
	}

	if entry.Statement != statement {
		t.Fatalf("expected statement %q, got %q", statement, entry.Statement)
	}
}
