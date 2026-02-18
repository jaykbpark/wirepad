package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVariables_Precedence(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".wirepad", "env"), 0o755); err != nil {
		t.Fatalf("mkdir private env: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "env"), 0o755); err != nil {
		t.Fatalf("mkdir shared env: %v", err)
	}

	writeFile(t, filepath.Join(root, ".env"), "source=dotenv\nbase=dotenv\n")
	writeFile(t, filepath.Join(root, "env", "dev.env"), "source=shared\nbase=shared\nshared_only=1\n")
	writeFile(t, filepath.Join(root, ".wirepad", "env", "dev.env"), "source=private\nbase=private\nprivate_only=1\n")

	opts := ResolveOptions{
		EnvName:      "dev",
		PrivateEnvDir: filepath.Join(root, ".wirepad", "env"),
		SharedEnvDir:  filepath.Join(root, "env"),
		DotEnvPath:    filepath.Join(root, ".env"),
		CLI: map[string]string{
			"source":   "cli",
			"cli_only": "1",
		},
	}

	got, err := ResolveVariables(opts)
	if err != nil {
		t.Fatalf("ResolveVariables returned error: %v", err)
	}

	if got["source"] != "cli" {
		t.Fatalf("expected CLI precedence for source, got %q", got["source"])
	}
	if got["base"] != "private" {
		t.Fatalf("expected private env precedence over shared/.env, got %q", got["base"])
	}
	if got["shared_only"] != "1" {
		t.Fatalf("expected shared env variable to be present")
	}
	if got["private_only"] != "1" {
		t.Fatalf("expected private env variable to be present")
	}
	if got["cli_only"] != "1" {
		t.Fatalf("expected CLI-only variable to be present")
	}
	if got["uuid"] == "" {
		t.Fatalf("expected generated uuid variable")
	}
	if got["timestamp_iso"] == "" {
		t.Fatalf("expected generated timestamp_iso variable")
	}
}

func TestResolveVariables_SkipsMissingFiles(t *testing.T) {
	got, err := ResolveVariables(ResolveOptions{
		EnvName: "dev",
		CLI: map[string]string{
			"foo": "bar",
		},
		PrivateEnvDir: filepath.Join(t.TempDir(), "missing"),
		SharedEnvDir:  filepath.Join(t.TempDir(), "missing"),
		DotEnvPath:    filepath.Join(t.TempDir(), "missing.env"),
	})
	if err != nil {
		t.Fatalf("ResolveVariables returned error: %v", err)
	}

	if got["foo"] != "bar" {
		t.Fatalf("expected foo=bar, got %q", got["foo"])
	}
}

func TestResolveVariables_InvalidEnvLine(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "broken-line\n")

	_, err := ResolveVariables(ResolveOptions{
		DotEnvPath: filepath.Join(root, ".env"),
	})
	if err == nil {
		t.Fatal("expected parse error for invalid env line")
	}
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
