package chalk_test

import (
	"errors"
	"testing"

	"github.com/chalk/chalk-go"
)

func TestOmittedLevelIsDetectedFromTheEnvironment(t *testing.T) {
	t.Setenv("FORCE_COLOR", "3")

	instance, err := chalk.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := instance.Level(); got != chalk.LevelTrueColor {
		t.Errorf("detected level: got %d, want %d", got, chalk.LevelTrueColor)
	}

	if got := instance.Red().Sprint("foo"); got != "\x1b[31mfoo\x1b[39m" {
		t.Errorf("styled output: got %q", got)
	}
}

// A nil Option is what a caller gets from a helper that decided not to
// configure anything, so it is skipped rather than treated as a failure.
func TestNilOptionsAreIgnored(t *testing.T) {
	instance, err := chalk.New(nil, chalk.WithLevel(chalk.LevelBasic), nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := instance.Level(); got != chalk.LevelBasic {
		t.Errorf("level: got %d, want %d", got, chalk.LevelBasic)
	}
}

func TestMustNewPanicsOnAnInvalidLevel(t *testing.T) {
	if got := chalk.MustNew(chalk.WithLevel(chalk.LevelAnsi256)).Level(); got != chalk.LevelAnsi256 {
		t.Errorf("level of a valid instance: got %d, want %d", got, chalk.LevelAnsi256)
	}

	defer func() {
		recovered := recover()

		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("recovered value: got %v, want an error", recovered)
		}

		if !errors.Is(err, chalk.ErrInvalidLevel) {
			t.Errorf("panic value: got %v, want it to wrap %v", err, chalk.ErrInvalidLevel)
		}
	}()

	chalk.MustNew(chalk.WithLevel(9))

	t.Error("MustNew with an invalid level returned instead of panicking")
}

// Level is the one method a caller can reach through a nil *Instance, because
// a zero Style renders through whatever instance it names.
func TestNilInstanceReportsNoColor(t *testing.T) {
	var instance *chalk.Instance

	if got := instance.Level(); got != chalk.LevelNone {
		t.Errorf("level of a nil instance: got %d, want %d", got, chalk.LevelNone)
	}
}
