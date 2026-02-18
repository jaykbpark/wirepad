package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecute_SendHTTPAndPersistRunHistory(t *testing.T) {
	withTempWorkingDir(t, func(root string) {
		var gotAuth string
		var gotName string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			defer r.Body.Close()

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			gotName = asString(body["name"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":"u_123","ok":true}`)
		}))
		defer server.Close()

		writeFile(t, filepath.Join(root, "requests", "users", "create.req.yaml"), `
version: 1
kind: http
name: users.create
request:
  method: POST
  url: "{{base_url}}/users"
  headers:
    Authorization: "Bearer {{token}}"
  body:
    mode: json
    json:
      name: "{{user_name}}"
`)
		writeFile(t, filepath.Join(root, "env", "dev.env"), "base_url="+server.URL+"\nuser_name=Alice\n")
		writeFile(t, filepath.Join(root, ".wirepad", "env", "dev.env"), "token=private-token\n")

		var out bytes.Buffer
		var errOut bytes.Buffer
		code := Execute([]string{"send", "users/create", "--env", "dev", "--var", "token=cli-token"}, &out, &errOut)

		if code != 0 {
			t.Fatalf("expected exit code 0, got %d; stderr=%q", code, errOut.String())
		}
		if errOut.Len() != 0 {
			t.Fatalf("expected no stderr output, got %q", errOut.String())
		}

		if gotAuth != "Bearer cli-token" {
			t.Fatalf("expected CLI var token to win, got auth header %q", gotAuth)
		}
		if gotName != "Alice" {
			t.Fatalf("expected interpolated user_name from env/dev.env, got %q", gotName)
		}

		stdout := out.String()
		if !strings.Contains(stdout, "POST 201 Created") {
			t.Fatalf("expected status line in stdout, got %q", stdout)
		}
		if !strings.Contains(stdout, "History: .wirepad/history/runs/") {
			t.Fatalf("expected history path in stdout, got %q", stdout)
		}

		entries, err := os.ReadDir(filepath.Join(root, ".wirepad", "history", "runs"))
		if err != nil {
			t.Fatalf("read run history directory: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 run history file, got %d", len(entries))
		}
	})
}

func withTempWorkingDir(t *testing.T, fn func(root string)) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir temp root: %v", err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	fn(root)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	data := strings.TrimLeft(content, "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func asString(value any) string {
	s, _ := value.(string)
	return s
}
