package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/looplj/graphql-cli/internal/audit"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Inspect recorded GraphQL operations",
	Long:  "Inspect the local audit log of executed GraphQL queries and mutations.",
}

var (
	auditListDetail   bool
	auditListEndpoint string
	auditListContains string
	auditListLimit    int
	auditListQuery    bool
	auditListMutation bool
	auditListStatus   string
)

func init() {
	auditCmd.AddCommand(newAuditListCmd())
	rootCmd.AddCommand(auditCmd)
}

func newAuditListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recorded GraphQL queries and mutations",
		Args:  cobra.NoArgs,
		RunE:  runAuditList,
	}

	cmd.Flags().BoolVar(&auditListDetail, "detail", false, "show the full GraphQL statement")
	cmd.Flags().StringVar(&auditListContains, "contains", "", "filter by text contained in statement, endpoint, URL, or error")
	cmd.Flags().StringVar(&auditListEndpoint, "endpoint", "", "filter by endpoint name")
	cmd.Flags().IntVar(&auditListLimit, "limit", 10, "limit results to the most recent N entries (0 = all)")
	cmd.Flags().BoolVar(&auditListQuery, "query", false, "show only queries")
	cmd.Flags().BoolVar(&auditListMutation, "mutation", false, "show only mutations")
	cmd.Flags().StringVar(&auditListStatus, "status", "", "filter by status: success or error")

	return cmd
}

func runAuditList(cmd *cobra.Command, args []string) error {
	if err := validateAuditStatus(auditListStatus); err != nil {
		return err
	}

	if auditListLimit < 0 {
		return fmt.Errorf("--limit must be zero or greater")
	}

	entries, err := audit.ReadEntries()
	if err != nil {
		return err
	}

	filtered := make([]audit.Entry, 0, len(entries))
	for _, entry := range entries {
		if !matchesAuditFilters(entry) {
			continue
		}

		filtered = append(filtered, entry)
	}

	if len(filtered) == 0 {
		fmt.Println("No audit entries found.")
		return nil
	}

	start := 0
	if auditListLimit > 0 && len(filtered) > auditListLimit {
		start = len(filtered) - auditListLimit
	}

	kindColor := color.New(color.FgCyan, color.Bold)
	successColor := color.New(color.FgGreen, color.Bold)
	errorColor := color.New(color.FgRed, color.Bold)
	dimColor := color.New(color.FgHiBlack)

	for i := len(filtered) - 1; i >= start; i-- {
		entry := filtered[i]
		kind := strings.ToUpper(audit.StatementKind(entry.Statement))

		kindColor.Printf("[%s] ", kind)
		fmt.Printf("%s ", entry.Endpoint)

		if entry.Status == "error" {
			errorColor.Printf("%s ", strings.ToUpper(entry.Status))
		} else {
			successColor.Printf("%s ", strings.ToUpper(entry.Status))
		}

		dimColor.Println(entry.Timestamp)

		statement := summarizeStatement(entry.Statement)
		if auditListDetail {
			statement = strings.TrimSpace(entry.Statement)
		}

		fmt.Printf("  %s\n", statement)

		if entry.Error != "" {
			errorColor.Printf("  error: %s\n", entry.Error)
		}

		if i > start {
			fmt.Println()
		}
	}

	return nil
}

func includeAuditKind(kind string) bool {
	if auditListQuery || auditListMutation {
		return (auditListQuery && kind == "query") || (auditListMutation && kind == "mutation")
	}

	return kind == "query" || kind == "mutation"
}

func matchesAuditFilters(entry audit.Entry) bool {
	if !includeAuditKind(audit.StatementKind(entry.Statement)) {
		return false
	}

	if auditListEndpoint != "" && entry.Endpoint != auditListEndpoint {
		return false
	}

	if auditListStatus != "" && !strings.EqualFold(entry.Status, auditListStatus) {
		return false
	}

	if auditListContains != "" && !auditEntryContains(entry, auditListContains) {
		return false
	}

	return true
}

func validateAuditStatus(status string) error {
	if status == "" {
		return nil
	}

	if strings.EqualFold(status, "success") || strings.EqualFold(status, "error") {
		return nil
	}

	return fmt.Errorf("invalid --status %q: must be success or error", status)
}

func auditEntryContains(entry audit.Entry, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	if needle == "" {
		return true
	}

	haystacks := []string{
		entry.Endpoint,
		entry.URL,
		entry.Status,
		entry.Statement,
		entry.Error,
	}

	for _, haystack := range haystacks {
		if strings.Contains(strings.ToLower(haystack), needle) {
			return true
		}
	}

	return false
}

func summarizeStatement(statement string) string {
	compact := strings.Join(strings.Fields(strings.TrimSpace(statement)), " ")
	if len(compact) <= 120 {
		return compact
	}

	return compact[:117] + "..."
}
