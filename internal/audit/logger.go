package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/looplj/graphql-cli/internal/config"
)

type Entry struct {
	Timestamp string `json:"timestamp"`
	Endpoint  string `json:"endpoint"`
	URL       string `json:"url,omitempty"`
	Status    string `json:"status"`
	Statement string `json:"statement"`
	Error     string `json:"error,omitempty"`
}

func LogExecution(ep *config.Endpoint, statement string, execErr error) error {
	if ep == nil {
		return fmt.Errorf("nil endpoint")
	}

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Endpoint:  ep.Name,
		URL:       ep.URL,
		Status:    "success",
		Statement: statement,
	}

	if execErr != nil {
		entry.Status = "error"
		entry.Error = execErr.Error()
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}

	path := config.DefaultAuditLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create audit log dir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}
