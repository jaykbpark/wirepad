package requestspec

import (
	"fmt"
	"os"
	"strings"
)

type LoadOptions struct {
	Strict bool
}

type LoadResult struct {
	Spec     *Spec
	Warnings []Issue
}

func LoadFile(path string, opts LoadOptions) (*LoadResult, error) {
	if !strings.HasSuffix(path, ".req.yaml") {
		return nil, fmt.Errorf("invalid request file %q: expected .req.yaml extension", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request file %q: %w", path, err)
	}

	spec, doc, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("load request file %q: %w", path, err)
	}

	issues := Validate(spec, doc, opts.Strict)

	var errs []Issue
	var warnings []Issue
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			errs = append(errs, issue)
			continue
		}
		warnings = append(warnings, issue)
	}

	if len(errs) > 0 {
		return nil, &ValidationError{
			Path:   path,
			Issues: errs,
		}
	}

	return &LoadResult{
		Spec:     spec,
		Warnings: warnings,
	}, nil
}
