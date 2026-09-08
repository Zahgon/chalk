package chalk

import (
	"io"

	"github.com/chalk/chalk-go/ansistyles"
)

// The style name lists, re-exported so callers do not need the ansistyles
// package for the common case. They mirror the named exports of the
// JavaScript entry point.
//
// Each list keeps the declaration order of the underlying table, because
// callers iterate them to build things like the rainbow example.
var (
	// ModifierNames lists the non-color styles, such as bold and
	// underlineCurly.
	ModifierNames = ansistyles.ModifierNames

	// ForegroundColorNames lists the named foreground colors.
	ForegroundColorNames = ansistyles.ForegroundColorNames

	// BackgroundColorNames lists the named background colors.
	BackgroundColorNames = ansistyles.BackgroundColorNames

	// UnderlineColorNames lists the named underline colors.
	//
	// These are deliberately absent from ColorNames, matching upstream.
	UnderlineColorNames = ansistyles.UnderlineColorNames

	// ColorNames lists the foreground and background color names together.
	ColorNames = ansistyles.ColorNames
)

// The functions below are the package level equivalent of calling the
// JavaScript `chalk` export directly, as in chalk('foo'). They render through
// [Default], so they follow its level.

// Sprint formats its operands and applies no styling.
//
// Operands are always joined with a single space, unlike [fmt.Sprint], which
// only inserts spaces between operands that are not strings. The JavaScript
// original joins with a space unconditionally.
func Sprint(a ...any) string {
	return Default.Root().Sprint(a...)
}

// Sprintf formats according to a format specifier.
func Sprintf(format string, a ...any) string {
	return Default.Root().Sprintf(format, a...)
}

// Sprintln formats its operands and appends a newline.
func Sprintln(a ...any) string {
	return Default.Root().Sprintln(a...)
}

// Print writes to os.Stdout.
func Print(a ...any) (int, error) {
	return Default.Root().Print(a...)
}

// Printf writes to os.Stdout according to a format specifier.
func Printf(format string, a ...any) (int, error) {
	return Default.Root().Printf(format, a...)
}

// Println writes to os.Stdout and appends a newline.
func Println(a ...any) (int, error) {
	return Default.Root().Println(a...)
}

// Fprint writes to w.
func Fprint(w io.Writer, a ...any) (int, error) {
	return Default.Root().Fprint(w, a...)
}

// Fprintf writes to w according to a format specifier.
func Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return Default.Root().Fprintf(w, format, a...)
}

// Fprintln writes to w and appends a newline.
func Fprintln(w io.Writer, a ...any) (int, error) {
	return Default.Root().Fprintln(w, a...)
}

// The same set on Instance, so that an instance can be called the way the
// JavaScript one can, as in chalkStderr('foo').

// Sprint formats its operands and applies no styling.
func (i *Instance) Sprint(a ...any) string {
	return i.Root().Sprint(a...)
}

// Sprintf formats according to a format specifier.
func (i *Instance) Sprintf(format string, a ...any) string {
	return i.Root().Sprintf(format, a...)
}

// Sprintln formats its operands and appends a newline.
func (i *Instance) Sprintln(a ...any) string {
	return i.Root().Sprintln(a...)
}

// Print writes to os.Stdout.
func (i *Instance) Print(a ...any) (int, error) {
	return i.Root().Print(a...)
}

// Printf writes to os.Stdout according to a format specifier.
func (i *Instance) Printf(format string, a ...any) (int, error) {
	return i.Root().Printf(format, a...)
}

// Println writes to os.Stdout and appends a newline.
func (i *Instance) Println(a ...any) (int, error) {
	return i.Root().Println(a...)
}

// Fprint writes to w.
func (i *Instance) Fprint(w io.Writer, a ...any) (int, error) {
	return i.Root().Fprint(w, a...)
}

// Fprintf writes to w according to a format specifier.
func (i *Instance) Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return i.Root().Fprintf(w, format, a...)
}

// Fprintln writes to w and appends a newline.
func (i *Instance) Fprintln(w io.Writer, a ...any) (int, error) {
	return i.Root().Fprintln(w, a...)
}
