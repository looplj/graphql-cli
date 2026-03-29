package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/looplj/graphql-cli/internal/config"
)

func TestReadEntriesParsesAuditLog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := config.DefaultAuditLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create audit dir: %v", err)
	}

	entries := []Entry{
		{
			Timestamp: "2026-03-30T01:02:03Z",
			Endpoint:  "production",
			Status:    "success",
			Statement: "{ viewer { id } }",
		},
		{
			Timestamp: "2026-03-30T01:03:04Z",
			Endpoint:  "production",
			Status:    "error",
			Statement: "# create a user\nmutation CreateUser { createUser(name: \"Amp\") { id } }",
			Error:     "HTTP 401",
		},
	}

	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}

		lines = append(lines, string(data))
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("write audit log: %v", err)
	}

	parsed, err := ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries returned error: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(parsed))
	}

	if got := StatementKind(parsed[0].Statement); got != "query" {
		t.Fatalf("expected first statement kind to be query, got %q", got)
	}

	if got := StatementKind(parsed[1].Statement); got != "mutation" {
		t.Fatalf("expected second statement kind to be mutation, got %q", got)
	}

	if parsed[1].Error != "HTTP 401" {
		t.Fatalf("expected error to round-trip, got %q", parsed[1].Error)
	}
}
