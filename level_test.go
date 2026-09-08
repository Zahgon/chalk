package chalk_test

import (
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chalk/chalk-go"
)

// findGoTool locates the go command. The toolchain that built this test is not
// guaranteed to be on PATH, so fall back to the GOROOT it was built with.
func findGoTool(t *testing.T) string {
	t.Helper()

	if tool, err := exec.LookPath("go"); err == nil {
		return tool
	}

	tool := filepath.Join(build.Default.GOROOT, "bin", "go")
	if _, err := os.Stat(tool); err != nil {
		t.Skipf("go tool not found: %v", err)
	}

	return tool
}

// Ported from test/level.js.

// withDefaultLevel runs fn with the default instance pinned to level, then
// restores the previous level.
//
// These tests mutate shared state the way the JavaScript ones do, so they
// must not run in parallel with anything that reads the default instance.
func withDefaultLevel(t *testing.T, level int, fn func()) {
	t.Helper()

	old := chalk.Default.Level()

	if err := chalk.Default.SetLevel(level); err != nil {
		t.Fatalf("SetLevel(%d): %v", level, err)
	}

	t.Cleanup(func() {
		if err := chalk.Default.SetLevel(old); err != nil {
			t.Fatalf("restore level: %v", err)
		}
	})

	fn()
}

func TestNoColorsWhenManuallyDisabled(t *testing.T) {
	withDefaultLevel(t, 0, func() {
		assertEqual(t, chalk.Red().Sprint("foo"), "foo")
	})
}

// A style captured from the default instance keeps reading that instance's
// level rather than a copy taken when it was created.
func TestEnableDisableColorsBasedOnOverallChalkLevelPropertyNotIndividualInstances(t *testing.T) {
	withDefaultLevel(t, 1, func() {
		red := chalk.Red()

		if got := red.Level(); got != 1 {
			t.Errorf("red level: got %d, want 1", got)
		}

		if err := chalk.Default.SetLevel(0); err != nil {
			t.Fatalf("SetLevel(0): %v", err)
		}

		if got, want := red.Level(), chalk.Default.Level(); got != want {
			t.Errorf("red level: got %d, want %d", got, want)
		}
	})
}

func TestPropagateChangesFromChildColors(t *testing.T) {
	withDefaultLevel(t, 1, func() {
		red := chalk.Red()

		if got := red.Level(); got != 1 {
			t.Errorf("red level: got %d, want 1", got)
		}

		if got := chalk.Default.Level(); got != 1 {
			t.Errorf("default level: got %d, want 1", got)
		}

		// The JavaScript test writes `red.level = 0`. A Go value type has
		// no assignable property, so the write goes through the instance
		// the style belongs to.
		if err := red.Instance().SetLevel(0); err != nil {
			t.Fatalf("red SetLevel(0): %v", err)
		}

		if got := red.Level(); got != 0 {
			t.Errorf("red level: got %d, want 0", got)
		}

		if got := chalk.Default.Level(); got != 0 {
			t.Errorf("default level: got %d, want 0", got)
		}

		if err := chalk.Default.SetLevel(1); err != nil {
			t.Fatalf("SetLevel(1): %v", err)
		}

		if got := red.Level(); got != 1 {
			t.Errorf("red level: got %d, want 1", got)
		}
	})
}

// The Go port of the JavaScript fixture test: run a real program with its
// output piped and no color-forcing variables, and check that detection
// disables color end to end.
func TestDisableColorsIfNotSupported(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess test in short mode")
	}

	cmd := exec.Command(findGoTool(t), "run", "./testdata/fixture")

	// A minimal environment, matching the extendEnv:false of the
	// JavaScript test, so the ambient TERM, CI, and COLORTERM cannot leak
	// in and enable color.
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"GOCACHE=" + os.Getenv("GOCACHE"),
		"GOMODCACHE=" + os.Getenv("GOMODCACHE"),
		"GOPATH=" + os.Getenv("GOPATH"),
	}

	// Output is captured through a pipe, so it is not a terminal.
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run fixture: %v", err)
	}

	if got := strings.TrimSpace(string(out)); got != "testout testerr" {
		t.Errorf("got %q, want %q", got, "testout testerr")
	}
}
