package chalk_test

import (
	"testing"

	"github.com/chalk/chalk-go"
)

// Ported from test/visible.js.

func TestVisibleNormalOutputWhenLevelAboveZero(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(3))

	assertEqual(t, instance.Visible().Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Red().Visible().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
}

func TestVisibleNoOutputWhenLevelIsTooLow(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(0))

	assertEqual(t, instance.Visible().Red().Sprint("foo"), "")
	assertEqual(t, instance.Red().Visible().Sprint("foo"), "")
}

// visible only suppresses output at level 0; everywhere else it is a no-op.
// Toggling the level back and forth checks that nothing is cached wrongly.
func TestSwitchingBetweenLevelZeroAndAbove(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(3))

	assertEqual(t, instance.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Visible().Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Red().Visible().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Visible().Sprint("foo"), "foo")
	assertEqual(t, instance.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")

	if err := instance.SetLevel(0); err != nil {
		t.Fatalf("SetLevel(0): %v", err)
	}

	assertEqual(t, instance.Red().Sprint("foo"), "foo")
	assertEqual(t, instance.Visible().Sprint("foo"), "")
	assertEqual(t, instance.Visible().Red().Sprint("foo"), "")
	assertEqual(t, instance.Red().Visible().Sprint("foo"), "")
	assertEqual(t, instance.Red().Sprint("foo"), "foo")

	if err := instance.SetLevel(3); err != nil {
		t.Fatalf("SetLevel(3): %v", err)
	}

	assertEqual(t, instance.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Visible().Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Red().Visible().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, instance.Visible().Sprint("foo"), "foo")
	assertEqual(t, instance.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
}
