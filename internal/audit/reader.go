package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/looplj/graphql-cli/internal/config"
)

func ReadEntries() ([]Entry, error) {
	file, err := os.Open(config.DefaultAuditLogPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close()

	entries := make([]Entry, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse audit log line %d: %w", lineNumber, err)
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}

	return entries, nil
}

func StatementKind(statement string) string {
	statement = trimGraphQLStatement(statement)
	lower := strings.ToLower(statement)

	switch {
	case lower == "":
		return "operation"
	case strings.HasPrefix(lower, "mutation"):
		return "mutation"
	case strings.HasPrefix(lower, "query"), strings.HasPrefix(lower, "{"):
		return "query"
	default:
		return "operation"
	}
}

func trimGraphQLStatement(statement string) string {
	statement = strings.TrimSpace(statement)
	for strings.HasPrefix(statement, "#") {
		newline := strings.IndexByte(statement, '\n')
		if newline == -1 {
			return ""
		}

		statement = strings.TrimSpace(statement[newline+1:])
	}

	return statement
}
