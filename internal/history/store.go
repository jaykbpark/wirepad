package history

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const runsDir = ".wirepad/history/runs"

type RunRecord struct {
	RunID           string            `json:"run_id"`
	RequestName     string            `json:"request_name"`
	RequestPath     string            `json:"request_path"`
	Env             string            `json:"env,omitempty"`
	StartedAt       string            `json:"started_at"`
	DurationMS      int64             `json:"duration_ms"`
	OK              bool              `json:"ok"`
	Status          int               `json:"status"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseBody    string            `json:"response_body,omitempty"`
}

func NewRunID(now time.Time) string {
	suffix := make([]byte, 2)
	if _, err := rand.Read(suffix); err != nil {
		return now.UTC().Format("2006-01-02T15-04-05Z")
	}
	return fmt.Sprintf("%s_%s", now.UTC().Format("2006-01-02T15-04-05Z"), hex.EncodeToString(suffix))
}

func SaveRun(record RunRecord) (string, error) {
	if record.RunID == "" {
		record.RunID = NewRunID(time.Now().UTC())
	}
	if record.StartedAt == "" {
		record.StartedAt = time.Now().UTC().Format(time.RFC3339)
	}

	if err := os.MkdirAll(runsDir, 0o755); err != nil {
		return "", fmt.Errorf("create history runs directory: %w", err)
	}

	path := filepath.Join(runsDir, record.RunID+".json")
	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode run record: %w", err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return "", fmt.Errorf("write run record: %w", err)
	}

	return path, nil
}
