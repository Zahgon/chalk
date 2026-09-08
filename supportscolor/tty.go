package supportscolor

import (
	"os"

	"golang.org/x/term"
)

// isTerminal reports whether f refers to a terminal.
//
// This is the equivalent of Node's tty.isatty on the stream's file descriptor.
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}

	return term.IsTerminal(int(f.Fd()))
}
