package cli

import "io"

func runSend(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printSendUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printSendUsage(stdout)
		return 0
	}
	return notImplemented(stderr, "wirepad send")
}

func printSendUsage(out io.Writer) {
	writeSimpleUsage(out, "wirepad send <request> [--env <name>] [--var key=value]")
}
