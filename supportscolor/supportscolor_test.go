package supportscolor_test

import (
	"testing"

	"github.com/chalk/chalk-go/supportscolor"
)

// These cases are ported from test/force-color.js in the JavaScript project.
//
// That suite spawns a subprocess per case with extendEnv:false so the ambient
// CI, TERM, and COLORTERM cannot leak in. Here the same isolation comes from
// passing an explicit Env map and Argv slice, which is both faster and
// stricter: a nil map would fall back to the real environment, so every case
// below passes a non-nil one.
//
// The JavaScript fixture writes to a pipe, so every case models a stream that
// exists but is not a terminal.
func detectLevel(env map[string]string, flags ...string) int {
	if env == nil {
		env = map[string]string{}
	}

	if flags == nil {
		flags = []string{}
	}

	support := supportscolor.Detect(supportscolor.Options{
		Env:        env,
		Argv:       flags,
		HaveStream: true,
		IsTTY:      false,
	})

	// A nil result is the Go spelling of the JavaScript `false`.
	if support == nil {
		return 0
	}

	return support.Level
}

func TestForceColorIsExactLevelNotMinimum(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want int
	}{
		{"FORCE_COLOR=1", map[string]string{"FORCE_COLOR": "1", "COLORTERM": "truecolor"}, 1},
		{"FORCE_COLOR=2", map[string]string{"FORCE_COLOR": "2", "COLORTERM": "truecolor"}, 2},
		{"FORCE_COLOR=3", map[string]string{"FORCE_COLOR": "3"}, 3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := detectLevel(c.env); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestForceColorOverridesDetection(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want int
	}{
		{"overrides CI detection", map[string]string{"FORCE_COLOR": "1", "CI": "true", "GITHUB_ACTIONS": "true"}, 1},
		{"overrides TERM detection", map[string]string{"FORCE_COLOR": "1", "TERM": "xterm-256color"}, 1},
		{"above 3 is clamped to 3", map[string]string{"FORCE_COLOR": "4", "TERM": "xterm-256color"}, 3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := detectLevel(c.env); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestColorFlagsTakePrecedenceOverNumericForceColor(t *testing.T) {
	if got := detectLevel(map[string]string{"FORCE_COLOR": "1"}, "--color=256"); got != 2 {
		t.Errorf("--color=256: got %d, want 2", got)
	}

	if got := detectLevel(map[string]string{"FORCE_COLOR": "1"}, "--color=16m"); got != 3 {
		t.Errorf("--color=16m: got %d, want 3", got)
	}
}

func TestForceColorZeroDisablesColor(t *testing.T) {
	if got := detectLevel(map[string]string{"FORCE_COLOR": "0", "COLORTERM": "truecolor"}); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// FORCE_COLOR=true switches color on but leaves the level to detection,
// unlike a numeric value which pins it.
func TestForceColorTrueOnlyEnablesColor(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want int
	}{
		{"with truecolor", map[string]string{"FORCE_COLOR": "true", "COLORTERM": "truecolor"}, 3},
		{"with xterm-256color", map[string]string{"FORCE_COLOR": "true", "TERM": "xterm-256color"}, 2},
		{"alone", map[string]string{"FORCE_COLOR": "true"}, 1},
		{"empty behaves like true", map[string]string{"FORCE_COLOR": "", "COLORTERM": "truecolor"}, 3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := detectLevel(c.env); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestForceColorFalseDisablesColor(t *testing.T) {
	if got := detectLevel(map[string]string{"FORCE_COLOR": "false", "COLORTERM": "truecolor"}); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// An unparseable FORCE_COLOR means "unset", not "disabled". Because the
// stream is piped both look like 0, so the Azure check -- which sits above
// the non-TTY check -- is what tells them apart.
func TestNonNumericForceColorIsUnsetNotDisabled(t *testing.T) {
	if got := detectLevel(map[string]string{"FORCE_COLOR": "unicorn", "COLORTERM": "truecolor"}); got != 0 {
		t.Errorf("piped: got %d, want 0", got)
	}

	azure := map[string]string{"FORCE_COLOR": "unicorn", "TF_BUILD": "1", "AGENT_NAME": "agent"}
	if got := detectLevel(azure); got != 1 {
		t.Errorf("azure: got %d, want 1", got)
	}

	disabled := map[string]string{"FORCE_COLOR": "0", "TF_BUILD": "1", "AGENT_NAME": "agent"}
	if got := detectLevel(disabled); got != 0 {
		t.Errorf("azure with FORCE_COLOR=0: got %d, want 0", got)
	}
}

func TestPartlyNumericForceColorIsUnsetNotALevel(t *testing.T) {
	for _, forceColor := range []string{" 2", "2 ", "2abc", "+2", "1e1", "0x2"} {
		t.Run(forceColor, func(t *testing.T) {
			piped := detectLevel(map[string]string{"FORCE_COLOR": forceColor, "COLORTERM": "truecolor"})
			if piped != 0 {
				t.Errorf("piped: got %d, want 0", piped)
			}

			azure := detectLevel(map[string]string{"FORCE_COLOR": forceColor, "TF_BUILD": "1", "AGENT_NAME": "agent"})
			if azure != 1 {
				t.Errorf("azure: got %d, want 1", azure)
			}
		})
	}
}

// The cases below are not in the JavaScript suite, which only exercises the
// FORCE_COLOR paths. They pin the rest of the ported decision tree so a
// reordering cannot pass unnoticed.

func detectTTY(env map[string]string, flags ...string) int {
	if flags == nil {
		flags = []string{}
	}

	support := supportscolor.Detect(supportscolor.Options{
		Env:        env,
		Argv:       flags,
		HaveStream: true,
		IsTTY:      true,
	})

	if support == nil {
		return 0
	}

	return support.Level
}

func TestDecisionTreeOnATerminal(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want int
	}{
		{"empty environment", map[string]string{}, 0},
		{"TERM=dumb wins over everything below it", map[string]string{"TERM": "dumb", "COLORTERM": "truecolor"}, 0},
		{"GitHub Actions", map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"}, 3},
		{"Gitea Actions", map[string]string{"CI": "true", "GITEA_ACTIONS": "true"}, 3},
		{"CircleCI", map[string]string{"CI": "true", "CIRCLECI": "true"}, 3},
		{"Travis", map[string]string{"CI": "true", "TRAVIS": "1"}, 1},
		{"AppVeyor", map[string]string{"CI": "true", "APPVEYOR": "1"}, 1},
		{"GitLab", map[string]string{"CI": "true", "GITLAB_CI": "1"}, 1},
		{"Buildkite", map[string]string{"CI": "true", "BUILDKITE": "1"}, 1},
		{"Drone", map[string]string{"CI": "true", "DRONE": "1"}, 1},
		{"Codeship", map[string]string{"CI": "true", "CI_NAME": "codeship"}, 1},
		{"unknown CI", map[string]string{"CI": "true"}, 0},
		{"TeamCity 9.1", map[string]string{"TEAMCITY_VERSION": "9.1.0"}, 1},
		{"TeamCity 10", map[string]string{"TEAMCITY_VERSION": "10.0.0"}, 1},
		{"TeamCity 9.0", map[string]string{"TEAMCITY_VERSION": "9.0.5"}, 0},
		{"COLORTERM=truecolor", map[string]string{"COLORTERM": "truecolor"}, 3},
		{"kitty", map[string]string{"TERM": "xterm-kitty"}, 3},
		{"ghostty", map[string]string{"TERM": "xterm-ghostty"}, 3},
		{"wezterm", map[string]string{"TERM": "wezterm"}, 3},
		{"iTerm 3", map[string]string{"TERM_PROGRAM": "iTerm.app", "TERM_PROGRAM_VERSION": "3.0.10"}, 3},
		{"iTerm 2", map[string]string{"TERM_PROGRAM": "iTerm.app", "TERM_PROGRAM_VERSION": "2.9.0"}, 2},
		{"Apple Terminal", map[string]string{"TERM_PROGRAM": "Apple_Terminal"}, 2},
		{"xterm-256color", map[string]string{"TERM": "xterm-256color"}, 2},
		{"screen-256", map[string]string{"TERM": "screen-256"}, 2},
		{"screen", map[string]string{"TERM": "screen"}, 1},
		{"xterm", map[string]string{"TERM": "xterm"}, 1},
		{"vt100", map[string]string{"TERM": "vt100"}, 1},
		{"rxvt", map[string]string{"TERM": "rxvt"}, 1},
		{"linux", map[string]string{"TERM": "linux"}, 1},
		{"cygwin", map[string]string{"TERM": "cygwin"}, 1},
		{"COLORTERM present but not truecolor", map[string]string{"COLORTERM": "1"}, 1},
		{"unknown TERM", map[string]string{"TERM": "unknown"}, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := detectTTY(c.env, "--"); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

// A non-TTY reports 0 unless something forces color, which is why almost
// every case above needs IsTTY.
func TestNonTTYReportsNoColor(t *testing.T) {
	if got := detectLevel(map[string]string{"COLORTERM": "truecolor"}); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// The Azure check deliberately sits above the non-TTY check.
func TestAzurePipelinesBeatsTheNonTTYCheck(t *testing.T) {
	if got := detectLevel(map[string]string{"TF_BUILD": "1", "AGENT_NAME": "agent"}); got != 1 {
		t.Errorf("got %d, want 1", got)
	}

	// Both variables are required.
	if got := detectLevel(map[string]string{"TF_BUILD": "1"}); got != 0 {
		t.Errorf("TF_BUILD alone: got %d, want 0", got)
	}
}

func TestFlagSniffing(t *testing.T) {
	cases := []struct {
		name  string
		flags []string
		want  int
	}{
		{"--color", []string{"--color"}, 1},
		{"--colors", []string{"--colors"}, 1},
		{"--color=true", []string{"--color=true"}, 1},
		{"--color=always", []string{"--color=always"}, 1},
		{"--color=16m", []string{"--color=16m"}, 3},
		{"--color=full", []string{"--color=full"}, 3},
		{"--color=truecolor", []string{"--color=truecolor"}, 3},
		{"--color=256", []string{"--color=256"}, 2},
		{"--no-color", []string{"--no-color"}, 0},
		{"--no-colors", []string{"--no-colors"}, 0},
		{"--color=false", []string{"--color=false"}, 0},
		{"--color=never", []string{"--color=never"}, 0},
		// Anything after a bare "--" is not a flag.
		{"after terminator", []string{"--", "--color"}, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := detectLevel(map[string]string{}, c.flags...); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestSkipFlagSniffing(t *testing.T) {
	support := supportscolor.Detect(supportscolor.Options{
		Env:              map[string]string{},
		Argv:             []string{"--color=16m"},
		HaveStream:       true,
		IsTTY:            false,
		SkipFlagSniffing: true,
	})

	if support != nil {
		t.Errorf("got level %d, want no support", support.Level)
	}
}

func TestTranslateLevel(t *testing.T) {
	// Level 0 is reported as a nil pointer, mirroring `false` in JavaScript.
	if got := supportscolor.Detect(supportscolor.Options{
		Env: map[string]string{"FORCE_COLOR": "0"}, Argv: []string{},
	}); got != nil {
		t.Fatalf("level 0: got %+v, want nil", got)
	}

	cases := []struct {
		level    string
		hasBasic bool
		has256   bool
		has16m   bool
	}{
		{"1", true, false, false},
		{"2", true, true, false},
		{"3", true, true, true},
	}

	for _, c := range cases {
		t.Run("FORCE_COLOR="+c.level, func(t *testing.T) {
			got := supportscolor.Detect(supportscolor.Options{
				Env: map[string]string{"FORCE_COLOR": c.level}, Argv: []string{},
			})

			if got == nil {
				t.Fatal("got nil, want support")
			}

			if got.HasBasic != c.hasBasic || got.Has256 != c.has256 || got.Has16m != c.has16m {
				t.Errorf("got %+v", got)
			}
		})
	}
}
