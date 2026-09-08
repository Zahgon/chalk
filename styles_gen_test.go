package chalk_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/chalk/chalk-go"
	"github.com/chalk/chalk-go/ansistyles"
)

// The generator emits three spellings of every entry in the ANSI style table:
// a method on Style, a method on *Instance, and a package-level function. The
// ported JavaScript tests only reach a handful of them, so a name wired to the
// wrong table entry, or a spelling the template forgot, would go unnoticed.
// These tests walk the table itself and check all three.

// packageStyles pins the package-level spelling of every table entry. It is
// written out by hand because Go cannot enumerate the functions of a package,
// which also makes it a second, independent list to compare the table against.
var packageStyles = map[string]func() chalk.Style{
	"reset":           chalk.Reset,
	"bold":            chalk.Bold,
	"dim":             chalk.Dim,
	"italic":          chalk.Italic,
	"underline":       chalk.Underline,
	"underlineDouble": chalk.UnderlineDouble,
	"underlineCurly":  chalk.UnderlineCurly,
	"underlineDotted": chalk.UnderlineDotted,
	"underlineDashed": chalk.UnderlineDashed,
	"overline":        chalk.Overline,
	"inverse":         chalk.Inverse,
	"hidden":          chalk.Hidden,
	"strikethrough":   chalk.Strikethrough,

	"black":         chalk.Black,
	"red":           chalk.Red,
	"green":         chalk.Green,
	"yellow":        chalk.Yellow,
	"blue":          chalk.Blue,
	"magenta":       chalk.Magenta,
	"cyan":          chalk.Cyan,
	"white":         chalk.White,
	"blackBright":   chalk.BlackBright,
	"gray":          chalk.Gray,
	"grey":          chalk.Grey,
	"redBright":     chalk.RedBright,
	"greenBright":   chalk.GreenBright,
	"yellowBright":  chalk.YellowBright,
	"blueBright":    chalk.BlueBright,
	"magentaBright": chalk.MagentaBright,
	"cyanBright":    chalk.CyanBright,
	"whiteBright":   chalk.WhiteBright,

	"bgBlack":         chalk.BgBlack,
	"bgRed":           chalk.BgRed,
	"bgGreen":         chalk.BgGreen,
	"bgYellow":        chalk.BgYellow,
	"bgBlue":          chalk.BgBlue,
	"bgMagenta":       chalk.BgMagenta,
	"bgCyan":          chalk.BgCyan,
	"bgWhite":         chalk.BgWhite,
	"bgBlackBright":   chalk.BgBlackBright,
	"bgGray":          chalk.BgGray,
	"bgGrey":          chalk.BgGrey,
	"bgRedBright":     chalk.BgRedBright,
	"bgGreenBright":   chalk.BgGreenBright,
	"bgYellowBright":  chalk.BgYellowBright,
	"bgBlueBright":    chalk.BgBlueBright,
	"bgMagentaBright": chalk.BgMagentaBright,
	"bgCyanBright":    chalk.BgCyanBright,
	"bgWhiteBright":   chalk.BgWhiteBright,

	"underlineBlack":         chalk.UnderlineBlack,
	"underlineRed":           chalk.UnderlineRed,
	"underlineGreen":         chalk.UnderlineGreen,
	"underlineYellow":        chalk.UnderlineYellow,
	"underlineBlue":          chalk.UnderlineBlue,
	"underlineMagenta":       chalk.UnderlineMagenta,
	"underlineCyan":          chalk.UnderlineCyan,
	"underlineWhite":         chalk.UnderlineWhite,
	"underlineBlackBright":   chalk.UnderlineBlackBright,
	"underlineGray":          chalk.UnderlineGray,
	"underlineGrey":          chalk.UnderlineGrey,
	"underlineRedBright":     chalk.UnderlineRedBright,
	"underlineGreenBright":   chalk.UnderlineGreenBright,
	"underlineYellowBright":  chalk.UnderlineYellowBright,
	"underlineBlueBright":    chalk.UnderlineBlueBright,
	"underlineMagentaBright": chalk.UnderlineMagentaBright,
	"underlineCyanBright":    chalk.UnderlineCyanBright,
	"underlineWhiteBright":   chalk.UnderlineWhiteBright,
}

// exportedName is the generator's naming rule, repeated here so the tests can
// derive the Go spelling of a table entry without importing the generator.
func exportedName(jsName string) string {
	return strings.ToUpper(jsName[:1]) + jsName[1:]
}

// chainMethod calls a no-argument method returning a Style, by name.
func chainMethod(t *testing.T, receiver any, name string) chalk.Style {
	t.Helper()

	method := reflect.ValueOf(receiver).MethodByName(name)
	if !method.IsValid() {
		t.Fatalf("%T has no method %s", receiver, name)
	}

	if got := method.Type().NumIn(); got != 0 {
		t.Fatalf("%T.%s takes %d arguments, want 0", receiver, name, got)
	}

	style, ok := method.Call(nil)[0].Interface().(chalk.Style)
	if !ok {
		t.Fatalf("%T.%s does not return a chalk.Style", receiver, name)
	}

	return style
}

// The hand-written list and the table must describe the same set of styles.
func TestPackageStyleListMatchesTheTable(t *testing.T) {
	if len(packageStyles) != len(ansistyles.All) {
		t.Errorf("packageStyles has %d entries, the style table has %d",
			len(packageStyles), len(ansistyles.All))
	}

	for name := range packageStyles {
		if _, ok := ansistyles.Lookup(name); !ok {
			t.Errorf("packageStyles has %q, which is not in the style table", name)
		}
	}
}

// Every table entry must be reachable under all three spellings, and all three
// must emit exactly the escape sequences recorded in the table.
func TestGeneratedStylesAgreeWithTheTable(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor))

	for _, style := range ansistyles.All {
		t.Run(style.Name, func(t *testing.T) {
			want := style.Open + "x" + style.Close
			goName := exportedName(style.Name)

			function, ok := packageStyles[style.Name]
			if !ok {
				t.Fatalf("no package-level function for %q", style.Name)
			}

			if got := function().Sprint("x"); got != want {
				t.Errorf("chalk.%s(): got %q, want %q", goName, got, want)
			}

			if got := chainMethod(t, chalk.Style{}, goName).Sprint("x"); got != want {
				t.Errorf("Style.%s(): got %q, want %q", goName, got, want)
			}

			if got := chainMethod(t, instance, goName).Sprint("x"); got != want {
				t.Errorf("(*Instance).%s(): got %q, want %q", goName, got, want)
			}
		})
	}
}

// A chain started on an instance must render at that instance's level, not at
// the level of the default instance the package-level functions use.
func TestGeneratedInstanceStylesFollowTheirOwnInstance(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(chalk.LevelNone))

	for _, style := range ansistyles.All {
		goName := exportedName(style.Name)

		if got := chainMethod(t, instance, goName).Sprint("x"); got != "x" {
			t.Errorf("(*Instance).%s() at level 0: got %q, want %q", goName, got, "x")
		}
	}

	if err := instance.SetLevel(chalk.LevelTrueColor); err != nil {
		t.Fatalf("SetLevel(3): %v", err)
	}

	if got, want := instance.Red().Sprint("x"), "\x1b[31mx\x1b[39m"; got != want {
		t.Errorf("(*Instance).Red() after SetLevel(3): got %q, want %q", got, want)
	}
}

// The parameterized styles are hand-written on Style and only forwarded by the
// generator, so the forwarding is what needs checking.
func TestGeneratedModelStylesForwardToStyle(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor))
	root := instance.Root()

	packageCases := []struct {
		name string
		got  string
		want string
	}{
		{"RGB", chalk.RGB(255, 0, 0).Sprint("x"), chalk.Style{}.RGB(255, 0, 0).Sprint("x")},
		{"BgRGB", chalk.BgRGB(255, 0, 0).Sprint("x"), chalk.Style{}.BgRGB(255, 0, 0).Sprint("x")},
		{"UnderlineRGB", chalk.UnderlineRGB(255, 0, 0).Sprint("x"), chalk.Style{}.UnderlineRGB(255, 0, 0).Sprint("x")},
		{"Hex", chalk.Hex("#FF0000").Sprint("x"), chalk.Style{}.Hex("#FF0000").Sprint("x")},
		{"BgHex", chalk.BgHex("#FF0000").Sprint("x"), chalk.Style{}.BgHex("#FF0000").Sprint("x")},
		{"UnderlineHex", chalk.UnderlineHex("#FF0000").Sprint("x"), chalk.Style{}.UnderlineHex("#FF0000").Sprint("x")},
		{"Ansi256", chalk.Ansi256(196).Sprint("x"), chalk.Style{}.Ansi256(196).Sprint("x")},
		{"BgAnsi256", chalk.BgAnsi256(196).Sprint("x"), chalk.Style{}.BgAnsi256(196).Sprint("x")},
		{"UnderlineAnsi256", chalk.UnderlineAnsi256(196).Sprint("x"), chalk.Style{}.UnderlineAnsi256(196).Sprint("x")},
	}

	for _, c := range packageCases {
		if c.got != c.want {
			t.Errorf("chalk.%s(): got %q, want %q", c.name, c.got, c.want)
		}
	}

	instanceCases := []struct {
		name string
		got  string
		want string
	}{
		{"RGB", instance.RGB(255, 0, 0).Sprint("x"), root.RGB(255, 0, 0).Sprint("x")},
		{"BgRGB", instance.BgRGB(255, 0, 0).Sprint("x"), root.BgRGB(255, 0, 0).Sprint("x")},
		{"UnderlineRGB", instance.UnderlineRGB(255, 0, 0).Sprint("x"), root.UnderlineRGB(255, 0, 0).Sprint("x")},
		{"Hex", instance.Hex("#FF0000").Sprint("x"), root.Hex("#FF0000").Sprint("x")},
		{"BgHex", instance.BgHex("#FF0000").Sprint("x"), root.BgHex("#FF0000").Sprint("x")},
		{"UnderlineHex", instance.UnderlineHex("#FF0000").Sprint("x"), root.UnderlineHex("#FF0000").Sprint("x")},
		{"Ansi256", instance.Ansi256(196).Sprint("x"), root.Ansi256(196).Sprint("x")},
		{"BgAnsi256", instance.BgAnsi256(196).Sprint("x"), root.BgAnsi256(196).Sprint("x")},
		{"UnderlineAnsi256", instance.UnderlineAnsi256(196).Sprint("x"), root.UnderlineAnsi256(196).Sprint("x")},
	}

	for _, c := range instanceCases {
		if c.got != c.want {
			t.Errorf("(*Instance).%s(): got %q, want %q", c.name, c.got, c.want)
		}
	}

	if got, want := instance.RGB(255, 0, 0).Sprint("x"), "\x1b[38;2;255;0;0mx\x1b[39m"; got != want {
		t.Errorf("(*Instance).RGB(255, 0, 0): got %q, want %q", got, want)
	}
}

// Visible and By are the two generated entry points that are not styles.
func TestGeneratedVisibleAndByStartAChain(t *testing.T) {
	instance := chalk.MustNew(chalk.WithLevel(chalk.LevelNone))

	if got := instance.Visible().Red().Sprint("x"); got != "" {
		t.Errorf("(*Instance).Visible() at level 0: got %q, want %q", got, "")
	}

	if got, want := chalk.Visible().Red().Sprint("x"), "\x1b[31mx\x1b[39m"; got != want {
		t.Errorf("chalk.Visible(): got %q, want %q", got, want)
	}

	style, err := instance.By("red")
	if err != nil {
		t.Fatalf("(*Instance).By(\"red\"): %v", err)
	}

	if got := style.Sprint("x"); got != "x" {
		t.Errorf("(*Instance).By(\"red\") at level 0: got %q, want %q", got, "x")
	}

	if _, err := instance.By("notAStyle"); err == nil {
		t.Error("(*Instance).By(\"notAStyle\"): got no error")
	}

	byRed, err := chalk.By("red")
	if err != nil {
		t.Fatalf("chalk.By(\"red\"): %v", err)
	}

	if got, want := byRed.Sprint("x"), "\x1b[31mx\x1b[39m"; got != want {
		t.Errorf("chalk.By(\"red\"): got %q, want %q", got, want)
	}
}
