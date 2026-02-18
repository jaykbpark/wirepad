package requestspec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFile_HTTPValid(t *testing.T) {
	path := writeRequestFile(t, "users-create.req.yaml", `
version: 1
kind: http
name: users.create
request:
  method: POST
  url: "https://api.example.com/users"
  body:
    mode: json
    json:
      email: "alice@example.com"
`)

	result, err := LoadFile(path, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(result.Warnings))
	}

	if result.Spec.Kind != KindHTTP {
		t.Fatalf("expected kind=%q, got %q", KindHTTP, result.Spec.Kind)
	}
}

func TestLoadFile_WSValid(t *testing.T) {
	path := writeRequestFile(t, "events-subscribe.req.yaml", `
version: 1
kind: ws
name: events.subscribe
request:
  url: "wss://api.example.com/events"
  messages:
    - type: text
      text: "hello"
`)

	result, err := LoadFile(path, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(result.Warnings))
	}

	if result.Spec.Kind != KindWS {
		t.Fatalf("expected kind=%q, got %q", KindWS, result.Spec.Kind)
	}
}

func TestLoadFile_HTTPMissingMethod(t *testing.T) {
	path := writeRequestFile(t, "missing-method.req.yaml", `
version: 1
kind: http
name: users.create
request:
  url: "https://api.example.com/users"
`)

	_, err := LoadFile(path, LoadOptions{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if !containsIssue(validationErr.Issues, "request.method", SeverityError, "missing required field") {
		t.Fatalf("expected request.method missing issue, got: %+v", validationErr.Issues)
	}
}

func TestLoadFile_UnknownFieldsWarnByDefault(t *testing.T) {
	path := writeRequestFile(t, "unknown-field.req.yaml", `
version: 1
kind: http
name: users.create
request:
  method: POST
  url: "https://api.example.com/users"
  mystery: true
surprise: value
`)

	result, err := LoadFile(path, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(result.Warnings))
	}

	if !containsIssue(result.Warnings, "request.mystery", SeverityWarning, "unknown field") {
		t.Fatalf("expected warning for request.mystery, got %+v", result.Warnings)
	}

	if !containsIssue(result.Warnings, "surprise", SeverityWarning, "unknown field") {
		t.Fatalf("expected warning for surprise, got %+v", result.Warnings)
	}
}

func TestLoadFile_UnknownFieldsErrorInStrictMode(t *testing.T) {
	path := writeRequestFile(t, "unknown-field-strict.req.yaml", `
version: 1
kind: http
name: users.create
request:
  method: POST
  url: "https://api.example.com/users"
  mystery: true
`)

	_, err := LoadFile(path, LoadOptions{Strict: true})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if !containsIssue(validationErr.Issues, "request.mystery", SeverityError, "unknown field") {
		t.Fatalf("expected strict-mode error for request.mystery, got %+v", validationErr.Issues)
	}
}

func TestLoadFile_InvalidBodyMode(t *testing.T) {
	path := writeRequestFile(t, "invalid-body-mode.req.yaml", `
version: 1
kind: http
name: users.create
request:
  method: POST
  url: "https://api.example.com/users"
  body:
    mode: jsno
`)

	_, err := LoadFile(path, LoadOptions{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if !containsIssue(validationErr.Issues, "request.body.mode", SeverityError, "invalid body mode") {
		t.Fatalf("expected invalid body mode issue, got %+v", validationErr.Issues)
	}
}

func TestLoadFile_RequiresReqYAMLExtension(t *testing.T) {
	path := writeRequestFile(t, "users-create.yaml", `
version: 1
kind: http
name: users.create
request:
  method: GET
  url: "https://api.example.com/users"
`)

	_, err := LoadFile(path, LoadOptions{})
	if err == nil {
		t.Fatal("expected extension error, got nil")
	}

	if !strings.Contains(err.Error(), ".req.yaml") {
		t.Fatalf("expected extension hint in error, got %v", err)
	}
}

func writeRequestFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write request file: %v", err)
	}

	return path
}

func containsIssue(issues []Issue, field string, severity Severity, message string) bool {
	for _, issue := range issues {
		if issue.Field == field && issue.Severity == severity && issue.Message == message {
			return true
		}
	}
	return false
}
