package cli

import (
	"fmt"
	"io"
)

func runWS(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printWSUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printWSUsage(stdout)
		return 0
	}

	switch args[0] {
	case "help":
		printWSUsage(stdout)
		return 0
	case "connect":
		return notImplemented(stderr, "wirepad ws connect")
	case "send":
		return notImplemented(stderr, "wirepad ws send")
	case "listen":
		return notImplemented(stderr, "wirepad ws listen")
	case "save-transcript":
		return notImplemented(stderr, "wirepad ws save-transcript")
	default:
		fmt.Fprintf(stderr, "unknown ws subcommand %q\n\n", args[0])
		printWSUsage(stderr)
		return 2
	}
}

func printWSUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  wirepad ws <subcommand> [args]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	fmt.Fprintln(out, "  connect          Open a WebSocket connection")
	fmt.Fprintln(out, "  send             Send frame(s) over an active WS connection")
	fmt.Fprintln(out, "  listen           Receive frames from an active WS connection")
	fmt.Fprintln(out, "  save-transcript  Save a WS transcript file")
}
