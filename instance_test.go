package chalk_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/chalk/chalk-go"
)

// Ported from test/instance.js.

func TestIsolatedContext(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(0))

	assertEqual(t, instance.Red().Sprint("foo"), "foo")
	assertEqual(t, chalk.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")

	if err := instance.SetLevel(2); err != nil {
		t.Fatalf("SetLevel(2): %v", err)
	}

	assertEqual(t, instance.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
}

func TestLevelOptionShouldBeANumberFrom0To3(t *testing.T) {
	for _, level := range []int{10, -1} {
		_, err := chalk.New(chalk.WithLevel(level))
		if err == nil {
			t.Errorf("WithLevel(%d): got no error", level)
			continue
		}

		// The JavaScript tests match on the message text, so the Go port
		// keeps the same wording.
		if !strings.Contains(err.Error(), "should be an integer from 0 to 3") {
			t.Errorf("WithLevel(%d): unexpected message %q", level, err)
		}

		if !errors.Is(err, chalk.ErrInvalidLevel) {
			t.Errorf("WithLevel(%d): does not wrap ErrInvalidLevel", level)
		}
	}
}

// Leaving the level out asks for detection. It must not be confused with
// asking for level 0, which is why the option is functional rather than a
// plain struct field.
func TestOmittedLevelIsDetectedNotRejected(t *testing.T) {
	a := chalk.MustNew()
	b := chalk.MustNew()

	if a.Level() != b.Level() {
		t.Errorf("detected levels differ: %d and %d", a.Level(), b.Level())
	}
}

func TestAssigningLevelIsValidated(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(1))

	// The JavaScript test also rejects 1.5, ' 1', and undefined. Those
	// cannot be expressed here: SetLevel takes an int, so the compiler
	// rejects them and a nil level is unrepresentable.
	for _, level := range []int{10, -1} {
		if err := instance.SetLevel(level); err == nil {
			t.Errorf("SetLevel(%d): got no error", level)
		}
	}

	// Validation applies to a style in the chain too, since it writes
	// through to the same instance.
	if err := instance.Red().Instance().SetLevel(10); err == nil {
		t.Error("Red().Instance().SetLevel(10): got no error")
	}

	// A rejected assignment must leave the level untouched.
	if got := instance.Level(); got != 1 {
		t.Errorf("level after failed assignments: got %d, want 1", got)
	}

	if err := instance.SetLevel(0); err != nil {
		t.Fatalf("SetLevel(0): %v", err)
	}

	assertEqual(t, instance.Red().Sprint("foo"), "foo")
}

// A model style resolves its escape sequence from the level in effect when it
// is created, so re-creating it after a level change downsamples it.
func TestModelStyleFollowsTheLevel(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(3))

	assertEqual(t, instance.RGB(255, 0, 0).Sprint("foo"), "\x1b[38;2;255;0;0mfoo\x1b[39m")

	if err := instance.SetLevel(1); err != nil {
		t.Fatalf("SetLevel(1): %v", err)
	}

	assertEqual(t, instance.RGB(255, 0, 0).Sprint("foo"), "\x1b[91mfoo\x1b[39m")

	if err := instance.SetLevel(0); err != nil {
		t.Fatalf("SetLevel(0): %v", err)
	}

	assertEqual(t, instance.RGB(255, 0, 0).Sprint("foo"), "foo")
}

func TestModelStyleOnAStyleInTheChainFollowsTheLevel(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(3))
	bold := instance.Bold()

	assertEqual(t, bold.RGB(255, 0, 0).Sprint("foo"),
		"\x1b[1m\x1b[38;2;255;0;0mfoo\x1b[39m\x1b[22m")

	if err := instance.SetLevel(1); err != nil {
		t.Fatalf("SetLevel(1): %v", err)
	}

	assertEqual(t, bold.RGB(255, 0, 0).Sprint("foo"),
		"\x1b[1m\x1b[91mfoo\x1b[39m\x1b[22m")
}

// Every style in a chain points back at the instance the chain started on,
// not at its immediate parent, so a level change is seen all the way down.
func TestDeepChainReadsTheLevelFromItsInstance(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(1))
	chain := instance.Red().Bold().Underline()

	if got := chain.Level(); got != 1 {
		t.Errorf("chain level: got %d, want 1", got)
	}

	if err := instance.SetLevel(0); err != nil {
		t.Fatalf("SetLevel(0): %v", err)
	}

	if got := chain.Level(); got != 0 {
		t.Errorf("chain level after change: got %d, want 0", got)
	}

	assertEqual(t, chain.Sprint("foo"), "foo")

	// Writing through the chain reaches the originating instance. In
	// JavaScript this is `chain.level = 2`; a Go value type cannot carry an
	// assignable property, so the write goes through Instance().
	if err := chain.Instance().SetLevel(2); err != nil {
		t.Fatalf("chain SetLevel(2): %v", err)
	}

	if got := instance.Level(); got != 2 {
		t.Errorf("instance level: got %d, want 2", got)
	}
}

func TestByRejectsUnknownStyleNames(t *testing.T) {
	if _, err := chalk.By("notAStyle"); err == nil {
		t.Error("By(\"notAStyle\"): got no error")
	}

	style, err := chalk.By("red")
	if err != nil {
		t.Fatalf("By(\"red\"): %v", err)
	}

	assertEqual(t, style.Sprint("foo"), "\x1b[31mfoo\x1b[39m")
}

// The zero Style must be usable, since Go gives no way to stop callers
// creating one.
func TestZeroStyleRendersThroughTheDefaultInstance(t *testing.T) {
	var zero chalk.Style

	assertEqual(t, zero.Sprint("foo"), "foo")
	assertEqual(t, zero.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
}
