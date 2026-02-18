package config

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultPrivateEnvDir = ".wirepad/env"
	defaultSharedEnvDir  = "env"
	defaultDotEnvPath    = ".env"
)

type ResolveOptions struct {
	EnvName       string
	CLI           map[string]string
	PrivateEnvDir string
	SharedEnvDir  string
	DotEnvPath    string
}

func ResolveVariables(opts ResolveOptions) (map[string]string, error) {
	privateEnvDir := opts.PrivateEnvDir
	if privateEnvDir == "" {
		privateEnvDir = defaultPrivateEnvDir
	}

	sharedEnvDir := opts.SharedEnvDir
	if sharedEnvDir == "" {
		sharedEnvDir = defaultSharedEnvDir
	}

	dotEnvPath := opts.DotEnvPath
	if dotEnvPath == "" {
		dotEnvPath = defaultDotEnvPath
	}

	resolved := generatedVars()

	// Lowest explicit file precedence: .env
	if err := mergeFileIfExists(resolved, dotEnvPath); err != nil {
		return nil, err
	}

	// Then shared env/<name>.env.
	if opts.EnvName != "" {
		if err := mergeFileIfExists(resolved, filepath.Join(sharedEnvDir, opts.EnvName+".env")); err != nil {
			return nil, err
		}
	}

	// Then private .wirepad/env/<name>.env.
	if opts.EnvName != "" {
		if err := mergeFileIfExists(resolved, filepath.Join(privateEnvDir, opts.EnvName+".env")); err != nil {
			return nil, err
		}
	}

	// Highest precedence: --var key=value.
	for key, value := range opts.CLI {
		resolved[key] = value
	}

	return resolved, nil
}

func mergeFileIfExists(dst map[string]string, path string) error {
	values, exists, err := parseEnvFile(path)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	for key, value := range values {
		dst[key] = value
	}
	return nil
}

func parseEnvFile(path string) (map[string]string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("open env file %q: %w", path, err)
	}
	defer f.Close()

	out := make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, true, fmt.Errorf("parse env file %q line %d: expected KEY=VALUE", path, lineNum)
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, true, fmt.Errorf("parse env file %q line %d: empty key", path, lineNum)
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		out[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, true, fmt.Errorf("read env file %q: %w", path, err)
	}

	return out, true, nil
}

func generatedVars() map[string]string {
	out := map[string]string{
		"timestamp_iso": time.Now().UTC().Format(time.RFC3339),
		"uuid":          newUUID(),
	}
	return out
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
