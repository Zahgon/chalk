package chalk_test

import (
	"testing"

	"github.com/chalk/chalk-go"
)

// Ported from test/no-color-support.js.

// The JavaScript test spoofs supports-color into reporting no color at all and
// then raises chalk.level by hand, proving that an explicit level wins over
// detection. Go cannot patch an imported module, so the detected level is
// overwritten directly, which is what the assignment does in JavaScript too.
func TestColorsCanBeForcedByUsingChalkLevel(t *testing.T) {
	withDefaultLevel(t, 0, func() {
		if got := chalk.Green().Sprint("hello"); got != "hello" {
			t.Fatalf("at level 0: got %q, want %q", got, "hello")
		}

		if err := chalk.Default.SetLevel(1); err != nil {
			t.Fatalf("SetLevel(1): %v", err)
		}

		if got, want := chalk.Green().Sprint("hello"), "\x1b[32mhello\x1b[39m"; got != want {
			t.Errorf("forced to level 1: got %q, want %q", got, want)
		}
	})
}
