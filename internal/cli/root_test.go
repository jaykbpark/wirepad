package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecute_NoArgs_ShowsRootUsage(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Execute(nil, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(out.String(), "wirepad <command> [args]") {
		t.Fatalf("expected root usage output, got %q", out.String())
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Execute([]string{"nope"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}

	if !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", errOut.String())
	}
}

func TestExecute_SendHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Execute([]string{"send", "--help"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(out.String(), "wirepad send <request>") {
		t.Fatalf("expected send usage, got %q", out.String())
	}
}

func TestExecute_SendMissingRequestFile(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Execute([]string{"send", "users/create"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	if !strings.Contains(errOut.String(), "resolve request:") {
		t.Fatalf("expected resolve request error, got %q", errOut.String())
	}
}
