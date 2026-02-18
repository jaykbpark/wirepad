package cli

import "io"

func runHist(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printHistUsage(stderr)
		return 2
	}
	if wantsHelp(args) {
		printHistUsage(stdout)
		return 0
	}
	return notImplemented(stderr, "wirepad hist")
}

func printHistUsage(out io.Writer) {
	writeSimpleUsage(out, "wirepad hist <request>")
}
