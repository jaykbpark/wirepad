package cli

import "io"

func runReplay(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printReplayUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printReplayUsage(stdout)
		return 0
	}
	return notImplemented(stderr, "wirepad replay")
}

func printReplayUsage(out io.Writer) {
	writeSimpleUsage(out, "wirepad replay <run_id>")
}
