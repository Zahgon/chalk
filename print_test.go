package chalk_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/chalk/chalk-go"
)

// JavaScript has one way to render a style: call it. Go splits that into the
// Sprint/Print/Fprint families, on Style, on *Instance, and at package level.
// The ported tests only reach Sprint, so everything else is checked here
// against the Sprint output it is defined in terms of.

// captureStdout redirects os.Stdout for the duration of fn.
//
// The print helpers resolve os.Stdout on each call, so swapping it is enough.
// Nothing in this suite runs in parallel, so the swap cannot be observed by
// another test.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	saved := os.Stdout
	os.Stdout = write

	// The pipe buffer is finite, so it has to be drained while fn writes.
	drained := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, read)
		drained <- buf.String()
	}()

	fn()

	os.Stdout = saved

	if err := write.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	out := <-drained

	if err := read.Close(); err != nil {
		t.Fatalf("close pipe reader: %v", err)
	}

	return out
}

func TestStyleRendersTheSameTextThroughEveryHelper(t *testing.T) {
	red := chalk.Red()
	want := "\x1b[31mfoo\x1b[39m"

	if got := red.Sprint("foo"); got != want {
		t.Errorf("Sprint: got %q, want %q", got, want)
	}

	if got := red.Sprintf("%s", "foo"); got != want {
		t.Errorf("Sprintf: got %q, want %q", got, want)
	}

	// The newline follows the closing sequence rather than being styled.
	if got := red.Sprintln("foo"); got != want+"\n" {
		t.Errorf("Sprintln: got %q, want %q", got, want+"\n")
	}

	// Sprintf formats first and styles the result, so a format string that
	// produces an empty string still produces no escape codes.
	if got := red.Sprintf("%s", ""); got != "" {
		t.Errorf("Sprintf with empty result: got %q, want %q", got, "")
	}

	var buf bytes.Buffer

	n, err := red.Fprint(&buf, "foo")
	if err != nil {
		t.Fatalf("Fprint: %v", err)
	}

	if n != len(want) {
		t.Errorf("Fprint wrote %d bytes, want %d", n, len(want))
	}

	if got := buf.String(); got != want {
		t.Errorf("Fprint: got %q, want %q", got, want)
	}

	buf.Reset()

	if _, err := red.Fprintf(&buf, "%s", "foo"); err != nil {
		t.Fatalf("Fprintf: %v", err)
	}

	if got := buf.String(); got != want {
		t.Errorf("Fprintf: got %q, want %q", got, want)
	}

	buf.Reset()

	if _, err := red.Fprintln(&buf, "foo"); err != nil {
		t.Fatalf("Fprintln: %v", err)
	}

	if got := buf.String(); got != want+"\n" {
		t.Errorf("Fprintln: got %q, want %q", got, want+"\n")
	}
}

func TestStylePrintWritesToStandardOutput(t *testing.T) {
	red := chalk.Red()
	want := "\x1b[31mfoo\x1b[39m"

	var (
		n   int
		err error
	)

	if got := captureStdout(t, func() { n, err = red.Print("foo") }); got != want {
		t.Errorf("Print: got %q, want %q", got, want)
	}

	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	if n != len(want) {
		t.Errorf("Print returned %d, want %d", n, len(want))
	}

	if got := captureStdout(t, func() { _, err = red.Printf("%s", "foo") }); got != want {
		t.Errorf("Printf: got %q, want %q", got, want)
	}

	if err != nil {
		t.Fatalf("Printf: %v", err)
	}

	if got := captureStdout(t, func() { _, err = red.Println("foo") }); got != want+"\n" {
		t.Errorf("Println: got %q, want %q", got, want+"\n")
	}

	if err != nil {
		t.Fatalf("Println: %v", err)
	}
}

// The package-level helpers are the Go spelling of calling `chalk` itself, so
// they apply no styling and follow the default instance.
func TestPackageHelpersRenderThroughTheDefaultInstance(t *testing.T) {
	if got := chalk.Sprintf("%s-%d", "foo", 1); got != "foo-1" {
		t.Errorf("Sprintf: got %q, want %q", got, "foo-1")
	}

	if got := chalk.Sprintln("foo", "bar"); got != "foo bar\n" {
		t.Errorf("Sprintln: got %q, want %q", got, "foo bar\n")
	}

	var buf bytes.Buffer

	if _, err := chalk.Fprint(&buf, "foo"); err != nil {
		t.Fatalf("Fprint: %v", err)
	}

	if _, err := chalk.Fprintf(&buf, "%s", "bar"); err != nil {
		t.Fatalf("Fprintf: %v", err)
	}

	if _, err := chalk.Fprintln(&buf, "baz"); err != nil {
		t.Fatalf("Fprintln: %v", err)
	}

	if got := buf.String(); got != "foobarbaz\n" {
		t.Errorf("Fprint/Fprintf/Fprintln: got %q, want %q", got, "foobarbaz\n")
	}

	var err error

	if got := captureStdout(t, func() { _, err = chalk.Print("foo") }); got != "foo" {
		t.Errorf("Print: got %q, want %q", got, "foo")
	}

	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	if got := captureStdout(t, func() { _, err = chalk.Printf("%s", "foo") }); got != "foo" {
		t.Errorf("Printf: got %q, want %q", got, "foo")
	}

	if err != nil {
		t.Fatalf("Printf: %v", err)
	}

	if got := captureStdout(t, func() { _, err = chalk.Println("foo") }); got != "foo\n" {
		t.Errorf("Println: got %q, want %q", got, "foo\n")
	}

	if err != nil {
		t.Fatalf("Println: %v", err)
	}
}

// The same helpers on an Instance are the Go spelling of `chalkStderr('foo')`.
// They must follow their own instance, which is what makes them worth having.
func TestInstanceHelpersRenderThroughTheirOwnInstance(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor))

	if got := instance.Sprint("foo", "bar"); got != "foo bar" {
		t.Errorf("Sprint: got %q, want %q", got, "foo bar")
	}

	if got := instance.Sprintf("%s-%d", "foo", 1); got != "foo-1" {
		t.Errorf("Sprintf: got %q, want %q", got, "foo-1")
	}

	if got := instance.Sprintln("foo"); got != "foo\n" {
		t.Errorf("Sprintln: got %q, want %q", got, "foo\n")
	}

	var buf bytes.Buffer

	if _, err := instance.Fprint(&buf, "foo"); err != nil {
		t.Fatalf("Fprint: %v", err)
	}

	if _, err := instance.Fprintf(&buf, "%s", "bar"); err != nil {
		t.Fatalf("Fprintf: %v", err)
	}

	if _, err := instance.Fprintln(&buf, "baz"); err != nil {
		t.Fatalf("Fprintln: %v", err)
	}

	if got := buf.String(); got != "foobarbaz\n" {
		t.Errorf("Fprint/Fprintf/Fprintln: got %q, want %q", got, "foobarbaz\n")
	}

	var err error

	if got := captureStdout(t, func() { _, err = instance.Print("foo") }); got != "foo" {
		t.Errorf("Print: got %q, want %q", got, "foo")
	}

	if err != nil {
		t.Fatalf("Print: %v", err)
	}

	if got := captureStdout(t, func() { _, err = instance.Printf("%s", "foo") }); got != "foo" {
		t.Errorf("Printf: got %q, want %q", got, "foo")
	}

	if err != nil {
		t.Fatalf("Printf: %v", err)
	}

	if got := captureStdout(t, func() { _, err = instance.Println("foo") }); got != "foo\n" {
		t.Errorf("Println: got %q, want %q", got, "foo\n")
	}

	if err != nil {
		t.Fatalf("Println: %v", err)
	}

	// The root style of an instance applies no styling, so the difference
	// between the instance helpers and the package ones is only the level.
	if err := instance.SetLevel(chalk.LevelNone); err != nil {
		t.Fatalf("SetLevel(0): %v", err)
	}

	if got := instance.Red().Sprint("foo"); got != "foo" {
		t.Errorf("Red().Sprint at level 0: got %q, want %q", got, "foo")
	}
}
