package cli

import (
	"fmt"
	"io"
)

func runReq(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printReqUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printReqUsage(stdout)
		return 0
	}

	switch args[0] {
	case "help":
		printReqUsage(stdout)
		return 0
	case "new":
		return notImplemented(stderr, "wirepad req new")
	case "edit":
		return notImplemented(stderr, "wirepad req edit")
	default:
		fmt.Fprintf(stderr, "unknown req subcommand %q\n\n", args[0])
		printReqUsage(stderr)
		return 2
	}
}

func printReqUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  wirepad req <subcommand> [args]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	fmt.Fprintln(out, "  new    Create a request spec file")
	fmt.Fprintln(out, "  edit   Edit and validate a request spec file")
}
