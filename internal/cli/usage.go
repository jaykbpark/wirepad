package cli

import (
	"fmt"
	"io"
)

func writeSimpleUsage(out io.Writer, usage string) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintf(out, "  %s\n", usage)
}
