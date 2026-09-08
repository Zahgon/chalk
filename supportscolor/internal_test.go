package supportscolor

import (
	"os"
	"testing"
)

// The exported behaviour is covered from supportscolor_test.go. These reach
// the unexported helpers directly, for inputs the public API cannot construct.

func TestArgvFallsBackToTheProcessArguments(t *testing.T) {
	saved := os.Args
	t.Cleanup(func() { os.Args = saved })

	os.Args = []string{"program", "--color", "--"}

	if got := (Options{}).argv(); len(got) != 2 || got[0] != "--color" {
		t.Errorf("argv with process arguments: got %q", got)
	}

	os.Args = []string{"program"}

	if got := (Options{}).argv(); got != nil {
		t.Errorf("argv with no process arguments: got %q, want nil", got)
	}
}

// hasFlag chooses the dash prefix from the length of the name it is given, so
// each spelling has to be exercised separately.
func TestHasFlagPicksThePrefixFromTheFlagSpelling(t *testing.T) {
	cases := []struct {
		flag string
		argv []string
		want bool
	}{
		{"color", []string{"--color"}, true},
		{"color", []string{"-color"}, false},
		{"c", []string{"-c"}, true},
		{"c", []string{"--c"}, false},
		{"--color", []string{"--color"}, true},
		{"-c", []string{"-c"}, true},
		{"--color", []string{"----color"}, false},
		{"color", []string{"--", "--color"}, false},
	}

	for _, c := range cases {
		if got := hasFlag(c.flag, c.argv); got != c.want {
			t.Errorf("hasFlag(%q, %q): got %t, want %t", c.flag, c.argv, got, c.want)
		}
	}
}

// A FORCE_COLOR of digits too long for an int is still a number far above the
// maximum level, so it must pin the level rather than be discarded.
func TestEnvForceColorClampsUnrepresentableNumbers(t *testing.T) {
	env := func(value string) lookup {
		return func(key string) (string, bool) {
			if key == "FORCE_COLOR" {
				return value, true
			}

			return "", false
		}
	}

	level, ok := envForceColor(env("99999999999999999999999999"))
	if !ok || level != LevelTrueColor {
		t.Errorf("an overlong number: got (%d, %t), want (%d, true)", level, ok, LevelTrueColor)
	}

	level, ok = envForceColor(env("2"))
	if !ok || level != LevelAnsi256 {
		t.Errorf("FORCE_COLOR=2: got (%d, %t), want (2, true)", level, ok)
	}
}

func TestIsTerminalRejectsAMissingFile(t *testing.T) {
	if isTerminal(nil) {
		t.Error("a nil file is not a terminal")
	}
}
