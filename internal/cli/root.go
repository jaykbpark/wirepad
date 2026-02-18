package cli

import (
	"fmt"
	"io"
)

// Execute dispatches CLI arguments and returns a process exit code.
func Execute(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stdout)
		return 0
	}

	cmd := args[0]
	if cmd == "-h" || cmd == "--help" {
		printRootUsage(stdout)
		return 0
	}
	if cmd == "help" {
		if len(args) == 1 {
			printRootUsage(stdout)
			return 0
		}
		return Execute([]string{args[1], "--help"}, stdout, stderr)
	}

	rest := args[1:]

	switch cmd {
	case "req":
		return runReq(rest, stdout, stderr)
	case "send":
		return runSend(rest, stdout, stderr)
	case "hist":
		return runHist(rest, stdout, stderr)
	case "diff":
		return runDiff(rest, stdout, stderr)
	case "replay":
		return runReplay(rest, stdout, stderr)
	case "ws":
		return runWS(rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", cmd)
		printRootUsage(stderr)
		return 2
	}
}

func printRootUsage(out io.Writer) {
	fmt.Fprintln(out, "wirepad: file-based API workflow CLI")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  wirepad <command> [args]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  req     Manage request specs")
	fmt.Fprintln(out, "  send    Execute HTTP request specs")
	fmt.Fprintln(out, "  hist    Show run history")
	fmt.Fprintln(out, "  diff    Compare run results")
	fmt.Fprintln(out, "  replay  Replay a previous run")
	fmt.Fprintln(out, "  ws      WebSocket workflows")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Run 'wirepad <command> --help' for details.")
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func notImplemented(stderr io.Writer, commandPath string) int {
	fmt.Fprintf(stderr, "%s is not implemented yet\n", commandPath)
	return 1
}
