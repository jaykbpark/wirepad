package config

import (
	"testing"

	"github.com/jaykbpark/wirepad/internal/requestspec"
)

func TestInterpolateAny_ReplacesSpecStrings(t *testing.T) {
	spec := &requestspec.Spec{
		Kind: requestspec.KindHTTP,
		Name: "users.create",
		Request: &requestspec.Request{
			Method: "POST",
			URL:    "{{base_url}}/users",
			Headers: map[string]any{
				"Authorization": "Bearer {{token}}",
			},
			Body: &requestspec.Body{
				Mode: "json",
				JSON: map[string]any{
					"email": "{{email}}",
				},
			},
		},
	}

	err := InterpolateAny(spec, map[string]string{
		"base_url": "https://api.example.com",
		"token":    "abc123",
		"email":    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("InterpolateAny returned error: %v", err)
	}

	if spec.Request.URL != "https://api.example.com/users" {
		t.Fatalf("unexpected URL: %q", spec.Request.URL)
	}

	if spec.Request.Headers["Authorization"] != "Bearer abc123" {
		t.Fatalf("unexpected header: %v", spec.Request.Headers["Authorization"])
	}

	bodyMap, ok := spec.Request.Body.JSON.(map[string]any)
	if !ok {
		t.Fatalf("expected json body to stay map[string]any")
	}

	if bodyMap["email"] != "alice@example.com" {
		t.Fatalf("unexpected json value: %v", bodyMap["email"])
	}
}

func TestInterpolateString_UnresolvedVariable(t *testing.T) {
	_, err := InterpolateString("{{missing}}", map[string]string{})
	if err == nil {
		t.Fatal("expected unresolved variable error")
	}
}
