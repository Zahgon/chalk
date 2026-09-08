package chalk_test

import (
	"os"
	"testing"

	"github.com/chalk/chalk-go"
)

// TestMain mirrors the preamble of the JavaScript test files, which pin the
// default instances to a known level so that the machine's real terminal
// support cannot change the expected output.
func TestMain(m *testing.M) {
	if err := chalk.Default.SetLevel(chalk.LevelTrueColor); err != nil {
		panic(err)
	}

	if err := chalk.Stderr.SetLevel(chalk.LevelTrueColor); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

// levelled returns the root style of a fresh instance, the Go spelling of
// `new Chalk({level})`.
func levelled(t *testing.T, level int) chalk.Style {
	t.Helper()

	instance, err := chalk.New(chalk.WithLevel(level))
	if err != nil {
		t.Fatalf("New(WithLevel(%d)): %v", level, err)
	}

	return instance.Root()
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Ported from test/chalk.js.

func TestNoStylingFromTheBaseFunction(t *testing.T) {
	assertEqual(t, chalk.Sprint("foo"), "foo")
}

func TestMultipleArgumentsInBaseFunction(t *testing.T) {
	assertEqual(t, chalk.Sprint("hello", "there"), "hello there")
}

func TestAutomaticCastingToString(t *testing.T) {
	assertEqual(t, chalk.Sprint(123), "123")
	assertEqual(t, chalk.Green().Sprint(98_765), "\x1b[32m98765\x1b[39m")

	// JavaScript coerces an array with join(','), giving 'hello,there'.
	// Go has no such coercion, so a slice formats the Go way. This is a
	// documented divergence; the assertion pins the Go behaviour.
	assertEqual(t, chalk.Sprint([]string{"hello", "there"}), "[hello there]")
	assertEqual(t, chalk.Bold().Sprint([]string{"foo", "bar"}), "\x1b[1m[foo bar]\x1b[22m")
}

func TestStyleString(t *testing.T) {
	assertEqual(t, chalk.Underline().Sprint("foo"), "\x1b[4mfoo\x1b[24m")
	assertEqual(t, chalk.Red().Sprint("foo"), "\x1b[31mfoo\x1b[39m")
	assertEqual(t, chalk.BgRed().Sprint("foo"), "\x1b[41mfoo\x1b[49m")
}

func TestApplyingMultipleStylesAtOnce(t *testing.T) {
	assertEqual(t,
		chalk.Red().BgGreen().Underline().Sprint("foo"),
		"\x1b[31m\x1b[42m\x1b[4mfoo\x1b[24m\x1b[49m\x1b[39m")

	assertEqual(t,
		chalk.Underline().Red().BgGreen().Sprint("foo"),
		"\x1b[4m\x1b[31m\x1b[42mfoo\x1b[49m\x1b[39m\x1b[24m")
}

func TestNestingStyles(t *testing.T) {
	inner := chalk.Underline().BgBlue().Sprint("bar")

	assertEqual(t,
		chalk.Red().Sprint("foo"+inner+"!"),
		"\x1b[31mfoo\x1b[4m\x1b[44mbar\x1b[49m\x1b[24m!\x1b[39m")
}

// Nesting two styles from the same group is the case that needs the close
// sequence to be rewritten back into the outer open sequence.
func TestNestingStylesOfTheSameType(t *testing.T) {
	nested := chalk.Yellow().Sprint("b" + chalk.Green().Sprint("c") + "b")

	assertEqual(t,
		chalk.Red().Sprint("a"+nested+"c"),
		"\x1b[31ma\x1b[33mb\x1b[32mc\x1b[39m\x1b[31m\x1b[33mb\x1b[39m\x1b[31mc\x1b[39m")
}

func TestResetAllStyles(t *testing.T) {
	inner := chalk.Red().BgGreen().Underline().Sprint("foo")

	assertEqual(t,
		chalk.Reset().Sprint(inner+"foo"),
		"\x1b[0m\x1b[31m\x1b[42m\x1b[4mfoo\x1b[24m\x1b[49m\x1b[39mfoo\x1b[0m")
}

func TestCachingMultipleStyles(t *testing.T) {
	red := chalk.Red().Red()
	green := chalk.Red().Green()
	redBold := red.Bold()
	greenBold := green.Bold()

	if red.Sprint("foo") == green.Sprint("foo") {
		t.Error("red and green produced the same output")
	}

	if redBold.Sprint("bar") == greenBold.Sprint("bar") {
		t.Error("redBold and greenBold produced the same output")
	}

	if green.Sprint("baz") == greenBold.Sprint("baz") {
		t.Error("green and greenBold produced the same output")
	}
}

func TestGrayAliasesGrey(t *testing.T) {
	assertEqual(t, chalk.Grey().Sprint("foo"), "\x1b[90mfoo\x1b[39m")
	assertEqual(t, chalk.Gray().Sprint("foo"), "\x1b[90mfoo\x1b[39m")
}

func TestVariableNumberOfArguments(t *testing.T) {
	assertEqual(t, chalk.Red().Sprint("foo", "bar"), "\x1b[31mfoo bar\x1b[39m")
}

// Zero is falsy in JavaScript but still produces output, unlike the empty
// string. The Go port has to reproduce that distinction.
func TestFalsyValues(t *testing.T) {
	assertEqual(t, chalk.Red().Sprint(0), "\x1b[31m0\x1b[39m")
}

func TestNoEscapeCodesForEmptyInput(t *testing.T) {
	assertEqual(t, chalk.Red().Sprint(), "")
	assertEqual(t, chalk.Red().Blue().Black().Sprint(), "")
	assertEqual(t, chalk.Red().Sprint(""), "")
}

func TestLineBreaksOpenAndCloseColors(t *testing.T) {
	assertEqual(t, chalk.Grey().Sprint("hello\nworld"),
		"\x1b[90mhello\x1b[39m\n\x1b[90mworld\x1b[39m")
}

func TestLineBreaksOpenAndCloseColorsWithCRLF(t *testing.T) {
	assertEqual(t, chalk.Grey().Sprint("hello\r\nworld"),
		"\x1b[90mhello\x1b[39m\r\n\x1b[90mworld\x1b[39m")
}

func TestConvertRGBToSixteenColors(t *testing.T) {
	one := levelled(t, 1)

	assertEqual(t, one.RGB(255, 0, 0).Sprint("hello"), "\x1b[91mhello\x1b[39m")
	assertEqual(t, one.BgRGB(255, 0, 0).Sprint("hello"), "\x1b[101mhello\x1b[49m")
	assertEqual(t, one.Hex("#FF0000").Sprint("hello"), "\x1b[91mhello\x1b[39m")
	assertEqual(t, one.BgHex("#FF0000").Sprint("hello"), "\x1b[101mhello\x1b[49m")
}

func TestConvertRGBTo256Colors(t *testing.T) {
	two, three := levelled(t, 2), levelled(t, 3)

	assertEqual(t, two.RGB(255, 0, 0).Sprint("hello"), "\x1b[38;5;196mhello\x1b[39m")
	assertEqual(t, two.BgRGB(255, 0, 0).Sprint("hello"), "\x1b[48;5;196mhello\x1b[49m")
	assertEqual(t, three.RGB(255, 0, 0).Sprint("hello"), "\x1b[38;2;255;0;0mhello\x1b[39m")
	assertEqual(t, three.BgRGB(255, 0, 0).Sprint("hello"), "\x1b[48;2;255;0;0mhello\x1b[49m")
	assertEqual(t, two.Hex("#FF0000").Sprint("hello"), "\x1b[38;5;196mhello\x1b[39m")
	assertEqual(t, two.BgHex("#FF0000").Sprint("hello"), "\x1b[48;5;196mhello\x1b[49m")
	assertEqual(t, three.BgHex("#FF0000").Sprint("hello"), "\x1b[48;2;255;0;0mhello\x1b[49m")
}

func TestConvertAnsi256ToSixteenColors(t *testing.T) {
	one := levelled(t, 1)

	assertEqual(t, one.Ansi256(196).Sprint("hello"), "\x1b[91mhello\x1b[39m")
	assertEqual(t, one.BgAnsi256(196).Sprint("hello"), "\x1b[101mhello\x1b[49m")
	assertEqual(t, one.Ansi256(2).Sprint("hello"), "\x1b[32mhello\x1b[39m")
	assertEqual(t, one.BgAnsi256(2).Sprint("hello"), "\x1b[42mhello\x1b[49m")
	assertEqual(t, one.Ansi256(8).Sprint("hello"), "\x1b[90mhello\x1b[39m")
	assertEqual(t, one.Ansi256(232).Sprint("hello"), "\x1b[30mhello\x1b[39m")
	assertEqual(t, one.Ansi256(255).Sprint("hello"), "\x1b[37mhello\x1b[39m")
}

func TestKeepAnsi256OnHigherLevels(t *testing.T) {
	two, three := levelled(t, 2), levelled(t, 3)

	assertEqual(t, two.Ansi256(196).Sprint("hello"), "\x1b[38;5;196mhello\x1b[39m")
	assertEqual(t, two.BgAnsi256(196).Sprint("hello"), "\x1b[48;5;196mhello\x1b[49m")
	assertEqual(t, three.Ansi256(196).Sprint("hello"), "\x1b[38;5;196mhello\x1b[39m")
	assertEqual(t, three.BgAnsi256(196).Sprint("hello"), "\x1b[48;5;196mhello\x1b[49m")
}

func TestNoColorCodesAtLevelZero(t *testing.T) {
	zero := levelled(t, 0)

	assertEqual(t, zero.Hex("#FF0000").Sprint("hello"), "hello")
	assertEqual(t, zero.BgHex("#FF0000").Sprint("hello"), "hello")
	assertEqual(t, zero.Ansi256(196).Sprint("hello"), "hello")
	assertEqual(t, zero.BgAnsi256(196).Sprint("hello"), "hello")
	assertEqual(t, zero.UnderlineHex("#FF0000").Sprint("hello"), "hello")
	assertEqual(t, zero.UnderlineAnsi256(196).Sprint("hello"), "hello")
	assertEqual(t, zero.UnderlineRed().Sprint("hello"), "hello")
	assertEqual(t, zero.UnderlineCurly().Sprint("hello"), "hello")
}

func TestExtendedUnderlineStyles(t *testing.T) {
	assertEqual(t, chalk.UnderlineDouble().Sprint("foo"), "\x1b[4:2mfoo\x1b[24m")
	assertEqual(t, chalk.UnderlineCurly().Sprint("foo"), "\x1b[4:3mfoo\x1b[24m")
	assertEqual(t, chalk.UnderlineDotted().Sprint("foo"), "\x1b[4:4mfoo\x1b[24m")
	assertEqual(t, chalk.UnderlineDashed().Sprint("foo"), "\x1b[4:5mfoo\x1b[24m")
}

func TestNestingUnderlineStyles(t *testing.T) {
	assertEqual(t,
		chalk.Underline().Sprint(chalk.UnderlineCurly().Sprint("a")+"b"),
		"\x1b[4m\x1b[4:3ma\x1b[24m\x1b[4mb\x1b[24m")

	assertEqual(t,
		chalk.UnderlineCurly().Sprint(chalk.Underline().Sprint("a")+"b"),
		"\x1b[4:3m\x1b[4ma\x1b[24m\x1b[4:3mb\x1b[24m")
}

func TestUnderlineColors(t *testing.T) {
	assertEqual(t, chalk.UnderlineRed().Sprint("foo"), "\x1b[58;5;1mfoo\x1b[59m")
	assertEqual(t, chalk.UnderlineBlackBright().Sprint("foo"), "\x1b[58;5;8mfoo\x1b[59m")
	assertEqual(t, chalk.UnderlineGray().Sprint("foo"), chalk.UnderlineBlackBright().Sprint("foo"))
	assertEqual(t, chalk.UnderlineGrey().Sprint("foo"), chalk.UnderlineBlackBright().Sprint("foo"))

	assertEqual(t,
		chalk.Red().UnderlineRed().UnderlineCurly().Sprint("foo"),
		"\x1b[31m\x1b[58;5;1m\x1b[4:3mfoo\x1b[24m\x1b[59m\x1b[39m")
}

func TestNestingUnderlineColors(t *testing.T) {
	assertEqual(t,
		chalk.UnderlineBlue().Sprint(chalk.UnderlineRed().Sprint("a")+"b"),
		"\x1b[58;5;4m\x1b[58;5;1ma\x1b[59m\x1b[58;5;4mb\x1b[59m")
}

func TestDownsampleUnderlineColors(t *testing.T) {
	one, two, three := levelled(t, 1), levelled(t, 2), levelled(t, 3)

	assertEqual(t, three.UnderlineRGB(255, 0, 0).Sprint("hello"), "\x1b[58;2;255;0;0mhello\x1b[59m")
	assertEqual(t, two.UnderlineRGB(255, 0, 0).Sprint("hello"), "\x1b[58;5;196mhello\x1b[59m")
	assertEqual(t, one.UnderlineRGB(255, 0, 0).Sprint("hello"), "\x1b[58;5;9mhello\x1b[59m")
	assertEqual(t, three.UnderlineHex("#FF0000").Sprint("hello"), "\x1b[58;2;255;0;0mhello\x1b[59m")
	assertEqual(t, two.UnderlineHex("#FF0000").Sprint("hello"), "\x1b[58;5;196mhello\x1b[59m")
	assertEqual(t, one.UnderlineHex("#FF0000").Sprint("hello"), "\x1b[58;5;9mhello\x1b[59m")
	assertEqual(t, three.UnderlineAnsi256(196).Sprint("hello"), "\x1b[58;5;196mhello\x1b[59m")
	assertEqual(t, two.UnderlineAnsi256(196).Sprint("hello"), "\x1b[58;5;196mhello\x1b[59m")
	assertEqual(t, one.UnderlineAnsi256(196).Sprint("hello"), "\x1b[58;5;9mhello\x1b[59m")
	assertEqual(t, one.UnderlineAnsi256(2).Sprint("hello"), "\x1b[58;5;2mhello\x1b[59m")
	assertEqual(t, one.UnderlineAnsi256(232).Sprint("hello"), "\x1b[58;5;0mhello\x1b[59m")

	// The named underline colors have no basic 16-color form, so they are
	// the same at every level.
	assertEqual(t, one.UnderlineRed().Sprint("hello"), "\x1b[58;5;1mhello\x1b[59m")
	assertEqual(t, two.UnderlineRed().Sprint("hello"), "\x1b[58;5;1mhello\x1b[59m")
}

func TestExposeUnderlineStyleNames(t *testing.T) {
	if !contains(chalk.ModifierNames, "underlineCurly") {
		t.Error("ModifierNames is missing underlineCurly")
	}

	if !contains(chalk.UnderlineColorNames, "underlineRedBright") {
		t.Error("UnderlineColorNames is missing underlineRedBright")
	}

	// Underline colors are intentionally not part of ColorNames.
	if contains(chalk.ColorNames, "underlineRed") {
		t.Error("ColorNames should not include underlineRed")
	}
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}

	return false
}

func TestBlackBrightColor(t *testing.T) {
	assertEqual(t, chalk.BlackBright().Sprint("foo"), "\x1b[90mfoo\x1b[39m")
}

func TestSetsCorrectLevelForChalkStderrAndRespectsIt(t *testing.T) {
	if got := chalk.Stderr.Level(); got != 3 {
		t.Fatalf("Stderr level: got %d, want 3", got)
	}

	assertEqual(t, chalk.Stderr.Red().Bold().Sprint("foo"),
		"\x1b[31m\x1b[1mfoo\x1b[22m\x1b[39m")
}

// The JavaScript styles are callable functions, so the test suite checks that
// Function.prototype's apply, bind, and call still work on them. A Go Style is
// a plain value, and the equivalent question is whether its rendering methods
// survive being detached from the chain: as a method value (bind), as a method
// expression applied to an explicit receiver (apply and call), and stored in a
// variable of function type.
func TestKeepFunctionPrototypeMethods(t *testing.T) {
	// Reflect.apply(chalk.grey, null, ['foo'])
	apply := chalk.Style.Sprint
	if got, want := apply(chalk.Grey(), "foo"), "\x1b[90mfoo\x1b[39m"; got != want {
		t.Errorf("method expression on grey: got %q, want %q", got, want)
	}

	// chalk.red.bgGreen.underline.bind(null)('foo')
	bound := chalk.Red().BgGreen().Underline().Sprint
	if got, want := chalk.Reset().Sprint(bound("foo")+"foo"),
		"\x1b[0m\x1b[31m\x1b[42m\x1b[4mfoo\x1b[24m\x1b[49m\x1b[39mfoo\x1b[0m"; got != want {
		t.Errorf("bound chain nested in reset: got %q, want %q", got, want)
	}

	// chalk.red.blue.black.call(null)
	if got := chalk.Red().Blue().Black().Sprint(); got != "" {
		t.Errorf("no operands: got %q, want an empty string", got)
	}
}

// The same question asked of the base style rather than of a chain: chalk
// itself is callable in JavaScript, and chalk.Sprint is detachable in Go.
func TestKeepsFunctionPrototypeMethods(t *testing.T) {
	root := chalk.Default.Root()

	// chalk.apply(chalk, ['foo'])
	if got := chalk.Style.Sprint(root, "foo"); got != "foo" {
		t.Errorf("method expression on the base style: got %q, want %q", got, "foo")
	}

	// chalk.bind(chalk, 'foo')()
	bound := func() string { return root.Sprint("foo") }
	if got := bound(); got != "foo" {
		t.Errorf("bound base style: got %q, want %q", got, "foo")
	}

	// chalk.call(chalk, 'foo')
	call := chalk.Sprint
	if got := call("foo"); got != "foo" {
		t.Errorf("package helper as a value: got %q, want %q", got, "foo")
	}
}
