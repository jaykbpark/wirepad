package cli

import "io"

func runDiff(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printDiffUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printDiffUsage(stdout)
		return 0
	}
	return notImplemented(stderr, "wirepad diff")
}

func printDiffUsage(out io.Writer) {
	writeSimpleUsage(out, "wirepad diff <request> [--last]")
}
