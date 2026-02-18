package requestspec

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath(ref string) (string, error) {
	ref = filepath.Clean(strings.TrimSpace(ref))
	if ref == "." || ref == "" {
		return "", fmt.Errorf("empty request reference")
	}

	if path, ok, err := existingFile(ref); err != nil {
		return "", err
	} else if ok {
		return path, nil
	}

	if strings.HasSuffix(ref, ".req.yaml") {
		return "", fmt.Errorf("request file %q not found", ref)
	}

	direct := filepath.Join("requests", filepath.FromSlash(ref)+".req.yaml")
	if path, ok, err := existingFile(direct); err != nil {
		return "", err
	} else if ok {
		return path, nil
	}

	needle := filepath.FromSlash(ref + ".req.yaml")
	var matches []string
	err := filepath.WalkDir("requests", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, needle) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("request %q not found (requests directory is missing)", ref)
		}
		return "", fmt.Errorf("walk requests directory: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("request %q not found", ref)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("request %q is ambiguous; matches: %s", ref, strings.Join(matches, ", "))
	}
	return matches[0], nil
}

func existingFile(path string) (string, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("stat %q: %w", path, err)
	}
	if info.IsDir() {
		return "", false, nil
	}
	return path, true, nil
}
