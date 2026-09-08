// Package supportscolor detects how much color a terminal supports.
//
// It is a port of the `supports-color` module vendored into chalk v6.0.0
// (source/vendor/supports-color/index.js). The decision order is reproduced
// exactly; several checks are deliberately ordered relative to the TTY test and
// to each other, and reordering them changes results on real terminals.
//
// The browser build of the original module has no meaningful Go equivalent and
// is not ported.
package supportscolor

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Color levels, matching chalk's levels.
const (
	// LevelNone means color is not supported.
	LevelNone = 0
	// LevelBasic means the 16 basic colors are supported.
	LevelBasic = 1
	// LevelAnsi256 means the 256-color palette is supported.
	LevelAnsi256 = 2
	// LevelTrueColor means 24-bit color is supported.
	LevelTrueColor = 3
)

// ColorSupport describes a stream's color capabilities.
//
// A nil *ColorSupport means color is not supported at all, mirroring the
// `false` the JavaScript returns for level 0.
type ColorSupport struct {
	// Level is the supported color level, from 1 to 3.
	Level int
	// HasBasic reports support for the 16 basic colors.
	HasBasic bool
	// Has256 reports support for the 256-color palette.
	Has256 bool
	// Has16m reports support for 24-bit color.
	Has16m bool
}

// Options configures [Detect].
type Options struct {
	// Env supplies environment variables. A nil map reads the process
	// environment.
	Env map[string]string

	// Argv supplies command line arguments to sniff. A nil slice reads
	// os.Args[1:].
	Argv []string

	// HaveStream reports whether the check concerns a real stream. When
	// false, the TTY check is skipped, matching a call with no stream.
	HaveStream bool

	// IsTTY reports whether that stream is a terminal.
	IsTTY bool

	// SkipFlagSniffing disables --color and --no-color sniffing.
	//
	// The zero value keeps sniffing enabled, matching the JavaScript default
	// of sniffFlags: true.
	SkipFlagSniffing bool
}

// lookup resolves environment variables, distinguishing unset from empty.
//
// Several checks turn on mere presence, so the distinction matters.
type lookup func(key string) (string, bool)

func (o Options) env() lookup {
	if o.Env == nil {
		return os.LookupEnv
	}

	return func(key string) (string, bool) {
		v, ok := o.Env[key]
		return v, ok
	}
}

func (o Options) argv() []string {
	if o.Argv == nil {
		if len(os.Args) > 1 {
			return os.Args[1:]
		}

		return nil
	}

	return o.Argv
}

// hasFlag reports whether a flag appears in argv before any "--" terminator.
//
// A flag is matched with the prefix its length implies: "-c" for a single
// character, "--color" otherwise. A flag given with its own leading dash is
// matched verbatim.
func hasFlag(flag string, argv []string) bool {
	prefix := "--"

	switch {
	case strings.HasPrefix(flag, "-"):
		prefix = ""
	case len(flag) == 1:
		prefix = "-"
	}

	target := prefix + flag

	for _, argument := range argv {
		// Everything after "--" belongs to the program being run.
		if argument == "--" {
			return false
		}

		if argument == target {
			return true
		}
	}

	return false
}

var digitsPattern = regexp.MustCompile(`^\d+$`)

// envForceColor interprets FORCE_COLOR.
//
// The second result reports whether FORCE_COLOR expressed an opinion. An unset
// variable and an unparseable value are both "no opinion", which is distinct
// from the explicit 0 that disables color.
func envForceColor(env lookup) (int, bool) {
	raw, ok := env("FORCE_COLOR")
	if !ok {
		return 0, false
	}

	switch raw {
	case "true", "":
		return LevelBasic, true
	case "false":
		return LevelNone, true
	}

	if !digitsPattern.MatchString(raw) {
		return 0, false
	}

	// A value too large for int is still a number far above the maximum level.
	n, err := strconv.Atoi(raw)
	if err != nil {
		return LevelTrueColor, true
	}

	return min(n, LevelTrueColor), true
}

// flagForceColor interprets --color and --no-color style flags.
func flagForceColor(argv []string) (int, bool) {
	if hasFlag("no-color", argv) ||
		hasFlag("no-colors", argv) ||
		hasFlag("color=false", argv) ||
		hasFlag("color=never", argv) {
		return LevelNone, true
	}

	if hasFlag("color", argv) ||
		hasFlag("colors", argv) ||
		hasFlag("color=true", argv) ||
		hasFlag("color=always", argv) {
		return LevelBasic, true
	}

	return 0, false
}

// Detect determines the color level for a stream.
//
// It returns nil when color is unsupported.
func Detect(opts Options) *ColorSupport {
	return translateLevel(detectLevel(opts))
}

// translateLevel converts a numeric level to a ColorSupport, using nil for
// "unsupported" the way the JavaScript uses false.
func translateLevel(level int) *ColorSupport {
	if level == LevelNone {
		return nil
	}

	return &ColorSupport{
		Level:    level,
		HasBasic: true,
		Has256:   level >= LevelAnsi256,
		Has16m:   level >= LevelTrueColor,
	}
}

// detectLevel is a direct port of _supportsColor. The order of the checks is
// significant throughout and is preserved from the source.
func detectLevel(opts Options) int {
	env := opts.env()
	argv := opts.argv()
	sniffFlags := !opts.SkipFlagSniffing

	envForce, envForceSet := envForceColor(env)

	// The JavaScript caches the flag-derived value in a module-level variable
	// and lets the environment overwrite it. Recomputing per call is
	// equivalent: the environment value is assigned before it is read, so the
	// cached value only survives when the environment has no opinion.
	force, forceSet := envForce, envForceSet
	if !forceSet && sniffFlags {
		force, forceSet = flagForceColor(argv)
	}

	if forceSet && force == LevelNone {
		return LevelNone
	}

	if sniffFlags {
		if hasFlag("color=16m", argv) ||
			hasFlag("color=full", argv) ||
			hasFlag("color=truecolor", argv) {
			return LevelTrueColor
		}

		if hasFlag("color=256", argv) {
			return LevelAnsi256
		}
	}

	// A numeric FORCE_COLOR pins the level outright, overriding detection.
	if forceSet {
		if raw, ok := env("FORCE_COLOR"); ok && digitsPattern.MatchString(raw) {
			return force
		}
	}

	// Azure DevOps is checked before the TTY test because it pipes output
	// while still rendering color.
	if _, ok := env("TF_BUILD"); ok {
		if _, ok := env("AGENT_NAME"); ok {
			return LevelBasic
		}
	}

	if opts.HaveStream && !opts.IsTTY && !forceSet {
		return LevelNone
	}

	// From here a forced level acts as a floor rather than an exact answer.
	minimum := LevelNone
	if forceSet {
		minimum = force
	}

	if term, _ := env("TERM"); term == "dumb" {
		return minimum
	}

	if level, ok := windowsLevel(); ok {
		return level
	}

	if _, ok := env("CI"); ok {
		for _, name := range []string{"GITHUB_ACTIONS", "GITEA_ACTIONS", "CIRCLECI"} {
			if _, ok := env(name); ok {
				return LevelTrueColor
			}
		}

		for _, name := range []string{"TRAVIS", "APPVEYOR", "GITLAB_CI", "BUILDKITE", "DRONE"} {
			if _, ok := env(name); ok {
				return LevelBasic
			}
		}

		if name, _ := env("CI_NAME"); name == "codeship" {
			return LevelBasic
		}

		return minimum
	}

	if version, ok := env("TEAMCITY_VERSION"); ok {
		if teamCityPattern.MatchString(version) {
			return LevelBasic
		}

		return LevelNone
	}

	if colorTerm, _ := env("COLORTERM"); colorTerm == "truecolor" {
		return LevelTrueColor
	}

	term, _ := env("TERM")

	switch term {
	case "xterm-kitty", "xterm-ghostty", "wezterm":
		return LevelTrueColor
	}

	if termProgram, ok := env("TERM_PROGRAM"); ok {
		rawVersion, _ := env("TERM_PROGRAM_VERSION")
		major, _, _ := strings.Cut(rawVersion, ".")
		version, _ := strconv.Atoi(major)

		switch termProgram {
		case "iTerm.app":
			if version >= 3 {
				return LevelTrueColor
			}

			return LevelAnsi256
		case "Apple_Terminal":
			return LevelAnsi256
		}
	}

	if term256Pattern.MatchString(term) {
		return LevelAnsi256
	}

	if termBasicPattern.MatchString(term) {
		return LevelBasic
	}

	if _, ok := env("COLORTERM"); ok {
		return LevelBasic
	}

	return minimum
}

var (
	// TeamCity gained ANSI support in 9.1.
	teamCityPattern = regexp.MustCompile(`^(?:9\.0*[1-9]\d*\.|\d{2,}\.)`)

	term256Pattern = regexp.MustCompile(`(?i)-256(?:color)?$`)

	// Only the first five alternatives are anchored; the rest match anywhere
	// in TERM. This asymmetry is present in the original pattern.
	termBasicPattern = regexp.MustCompile(`(?i)^screen|^xterm|^vt100|^vt220|^rxvt|color|ansi|cygwin|linux`)
)

// Stdout reports the color support of standard output.
func Stdout() *ColorSupport {
	return Detect(Options{HaveStream: true, IsTTY: isTerminal(os.Stdout)})
}

// Stderr reports the color support of standard error.
func Stderr() *ColorSupport {
	return Detect(Options{HaveStream: true, IsTTY: isTerminal(os.Stderr)})
}
