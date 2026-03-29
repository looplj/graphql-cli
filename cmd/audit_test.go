package cmd

import (
	"testing"

	"github.com/looplj/graphql-cli/internal/audit"
)

func TestMatchesAuditFilters(t *testing.T) {
	entry := audit.Entry{
		Endpoint:  "production",
		URL:       "https://api.example.com/graphql",
		Status:    "error",
		Statement: "mutation CreateUser { createUser(name: \"Amp\") { id } }",
		Error:     "HTTP 401",
	}

	auditListEndpoint = "production"
	auditListMutation = true
	auditListQuery = false
	auditListStatus = "error"
	auditListContains = "createuser"

	if !matchesAuditFilters(entry) {
		t.Fatal("expected entry to match combined filters")
	}

	auditListContains = "viewer"

	if matchesAuditFilters(entry) {
		t.Fatal("expected entry not to match when contains filter misses")
	}

	auditListContains = ""
	auditListStatus = "success"

	if matchesAuditFilters(entry) {
		t.Fatal("expected entry not to match mismatched status")
	}
}

func TestValidateAuditStatus(t *testing.T) {
	if err := validateAuditStatus("success"); err != nil {
		t.Fatalf("expected success to be valid: %v", err)
	}

	if err := validateAuditStatus("error"); err != nil {
		t.Fatalf("expected error to be valid: %v", err)
	}

	if err := validateAuditStatus("pending"); err == nil {
		t.Fatal("expected invalid status to return an error")
	}
}
