package requestspec

import (
	"fmt"
	"slices"
	"strings"
)

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Issue struct {
	Severity Severity
	Field    string
	Message  string
	Hint     string
}

type ValidationError struct {
	Path   string
	Issues []Issue
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	if e.Path != "" {
		fmt.Fprintf(&b, "validation error in %s", e.Path)
	} else {
		b.WriteString("validation error")
	}

	for _, issue := range e.Issues {
		fmt.Fprintf(&b, "\n- %s: %s", issue.Field, issue.Message)
		if issue.Hint != "" {
			fmt.Fprintf(&b, " (hint: %s)", issue.Hint)
		}
	}

	return b.String()
}

func Validate(spec *Spec, raw map[string]any, strict bool) []Issue {
	var issues []Issue

	add := func(severity Severity, field, message, hint string) {
		issues = append(issues, Issue{
			Severity: severity,
			Field:    field,
			Message:  message,
			Hint:     hint,
		})
	}

	if spec.Version == 0 {
		add(SeverityError, "version", "missing required field", "set version: 1")
	} else if spec.Version != 1 {
		add(SeverityError, "version", "unsupported version", "version: 1 is currently supported")
	}

	if strings.TrimSpace(string(spec.Kind)) == "" {
		add(SeverityError, "kind", "missing required field", "expected one of [http, ws]")
	} else if spec.Kind != KindHTTP && spec.Kind != KindWS {
		add(SeverityError, "kind", "invalid kind", "expected one of [http, ws]")
	}

	if strings.TrimSpace(spec.Name) == "" {
		add(SeverityError, "name", "missing required field", "set a stable request name")
	}

	if spec.Request == nil {
		add(SeverityError, "request", "missing required field", "define request block")
	} else {
		if strings.TrimSpace(spec.Request.URL) == "" {
			add(SeverityError, "request.url", "missing required field", "set a request URL")
		}

		if spec.Kind == KindHTTP && strings.TrimSpace(spec.Request.Method) == "" {
			add(SeverityError, "request.method", "missing required field", "kind=http requires request.method")
		}

		validateBody(spec.Request.Body, add)
		validateWSMessages(spec.Request.Messages, add)
	}

	for _, issue := range unknownFieldIssues(raw, strict) {
		add(issue.Severity, issue.Field, issue.Message, issue.Hint)
	}

	return issues
}

func validateBody(body *Body, add func(Severity, string, string, string)) {
	if body == nil {
		return
	}

	if strings.TrimSpace(body.Mode) == "" {
		add(SeverityError, "request.body.mode", "missing required field", "expected one of [json, raw, file, form, multipart]")
		return
	}

	validModes := []string{"json", "raw", "file", "form", "multipart"}
	if !slices.Contains(validModes, body.Mode) {
		add(SeverityError, "request.body.mode", "invalid body mode", "expected one of [json, raw, file, form, multipart]")
		return
	}

	switch body.Mode {
	case "json":
		if body.JSON == nil {
			add(SeverityError, "request.body.json", "missing payload for json mode", "set request.body.json")
		}
	case "raw":
		if strings.TrimSpace(body.Raw) == "" {
			add(SeverityError, "request.body.raw", "missing payload for raw mode", "set request.body.raw")
		}
	case "file":
		if strings.TrimSpace(body.Path) == "" {
			add(SeverityError, "request.body.path", "missing file path for file mode", "set request.body.path")
		}
	case "form":
		if len(body.Form) == 0 {
			add(SeverityError, "request.body.form", "missing fields for form mode", "set request.body.form")
		}
	case "multipart":
		if len(body.Multipart) == 0 {
			add(SeverityError, "request.body.multipart", "missing parts for multipart mode", "set request.body.multipart")
		}
	}
}

func validateWSMessages(messages []WSMessage, add func(Severity, string, string, string)) {
	validTypes := []string{"json", "text", "file"}
	for i, msg := range messages {
		fieldPrefix := fmt.Sprintf("request.messages[%d]", i)

		if strings.TrimSpace(msg.Type) == "" {
			add(SeverityError, fieldPrefix+".type", "missing required field", "expected one of [json, text, file]")
			continue
		}

		if !slices.Contains(validTypes, msg.Type) {
			add(SeverityError, fieldPrefix+".type", "invalid message type", "expected one of [json, text, file]")
			continue
		}

		switch msg.Type {
		case "json":
			if msg.JSON == nil {
				add(SeverityError, fieldPrefix+".json", "missing payload for json message", "set "+fieldPrefix+".json")
			}
		case "text":
			if strings.TrimSpace(msg.Text) == "" {
				add(SeverityError, fieldPrefix+".text", "missing payload for text message", "set "+fieldPrefix+".text")
			}
		case "file":
			if strings.TrimSpace(msg.Path) == "" {
				add(SeverityError, fieldPrefix+".path", "missing file path for file message", "set "+fieldPrefix+".path")
			}
		}
	}
}

type schemaNode struct {
	children map[string]*schemaNode
	elem     *schemaNode
	allowAny bool
}

func unknownFieldIssues(raw map[string]any, strict bool) []Issue {
	var issues []Issue
	if raw == nil {
		return issues
	}

	root := requestFileSchema()
	walkUnknownFields(raw, root, "", strict, &issues)
	return issues
}

func walkUnknownFields(node any, schema *schemaNode, path string, strict bool, issues *[]Issue) {
	if node == nil || schema == nil || schema.allowAny {
		return
	}

	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			field := joinPath(path, key)

			child, ok := schema.children[key]
			if !ok {
				severity := SeverityWarning
				if strict {
					severity = SeverityError
				}
				*issues = append(*issues, Issue{
					Severity: severity,
					Field:    field,
					Message:  "unknown field",
					Hint:     "remove field or run without --strict",
				})
				continue
			}

			walkUnknownFields(value, child, field, strict, issues)
		}
	case []any:
		if schema.elem == nil {
			return
		}
		for _, child := range typed {
			walkUnknownFields(child, schema.elem, path, strict, issues)
		}
	}
}

func requestFileSchema() *schemaNode {
	anyMap := &schemaNode{allowAny: true}

	messageSchema := &schemaNode{
		children: map[string]*schemaNode{
			"type": scalarSchema(),
			"json": anyMap,
			"text": scalarSchema(),
			"path": scalarSchema(),
		},
	}

	bodySchema := &schemaNode{
		children: map[string]*schemaNode{
			"mode":         scalarSchema(),
			"json":         anyMap,
			"raw":          scalarSchema(),
			"path":         scalarSchema(),
			"content_type": scalarSchema(),
			"form":         anyMap,
			"multipart":    sequenceSchema(anyMap),
		},
	}

	requestSchema := &schemaNode{
		children: map[string]*schemaNode{
			"method":             scalarSchema(),
			"url":                scalarSchema(),
			"query":              anyMap,
			"headers":            anyMap,
			"body":               bodySchema,
			"timeout_ms":         scalarSchema(),
			"follow_redirects":   scalarSchema(),
			"connect_timeout_ms": scalarSchema(),
			"ping_interval_ms":   scalarSchema(),
			"messages":           sequenceSchema(messageSchema),
		},
	}

	return &schemaNode{
		children: map[string]*schemaNode{
			"version":     scalarSchema(),
			"kind":        scalarSchema(),
			"name":        scalarSchema(),
			"description": scalarSchema(),
			"tags":        sequenceSchema(scalarSchema()),
			"request":     requestSchema,
			"expect":      anyMap,
			"hooks":       anyMap,
		},
	}
}

func scalarSchema() *schemaNode {
	return &schemaNode{}
}

func sequenceSchema(elem *schemaNode) *schemaNode {
	return &schemaNode{elem: elem}
}

func joinPath(base, segment string) string {
	if base == "" {
		return segment
	}
	return base + "." + segment
}
