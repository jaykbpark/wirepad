package requestspec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath_ExactPath(t *testing.T) {
	withTempWorkingDir(t, func(root string) {
		path := filepath.Join(root, "requests", "users", "create.req.yaml")
		writeFile(t, path, "version: 1\nkind: http\nname: users.create\nrequest:\n  method: GET\n  url: https://example.com\n")

		got, err := ResolvePath(path)
		if err != nil {
			t.Fatalf("ResolvePath returned error: %v", err)
		}
		if got != path {
			t.Fatalf("expected %q, got %q", path, got)
		}
	})
}

func TestResolvePath_ByName(t *testing.T) {
	withTempWorkingDir(t, func(root string) {
		path := filepath.Join("requests", "users", "create.req.yaml")
		writeFile(t, filepath.Join(root, path), "version: 1\nkind: http\nname: users.create\nrequest:\n  method: GET\n  url: https://example.com\n")

		got, err := ResolvePath("users/create")
		if err != nil {
			t.Fatalf("ResolvePath returned error: %v", err)
		}
		if got != path {
			t.Fatalf("expected %q, got %q", path, got)
		}
	})
}

func TestResolvePath_NotFound(t *testing.T) {
	withTempWorkingDir(t, func(string) {
		if _, err := ResolvePath("users/missing"); err == nil {
			t.Fatal("expected not found error")
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
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
