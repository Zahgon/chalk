package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/chalk/chalk-go"
)

// wantScreenshot is the verbatim output of `FORCE_COLOR=3 node examples/screenshot.js`
// captured from the JavaScript implementation.
const wantScreenshot = "\x1b[1mbold\x1b[22m \x1b[2mdim\x1b[22m \x1b[3mitalic\x1b[23m " +
	"\x1b[4munderline\x1b[24m \x1b[7minverse\x1b[27m \x1b[9mstrikethrough\x1b[29m " +
	"\x1b[30mblack\x1b[39m \x1b[31mred\x1b[39m \x1b[32mgreen\x1b[39m \x1b[33myellow\x1b[39m " +
	"\x1b[34mblue\x1b[39m \x1b[35mmagenta\x1b[39m \x1b[36mcyan\x1b[39m \x1b[37mwhite\x1b[39m " +
	"\x1b[90mgray\x1b[39m \x1b[40mbgBlack\x1b[49m \x1b[41m\x1b[30mbgRed\x1b[39m\x1b[49m " +
	"\x1b[42m\x1b[30mbgGreen\x1b[39m\x1b[49m \x1b[43m\x1b[30mbgYellow\x1b[39m\x1b[49m " +
	"\x1b[44mbgBlue\x1b[49m \x1b[45m\x1b[30mbgMagenta\x1b[39m\x1b[49m " +
	"\x1b[46m\x1b[30mbgCyan\x1b[39m\x1b[49m \x1b[47m\x1b[30mbgWhite\x1b[39m\x1b[49m "

func TestRenderMatchesJavaScript(t *testing.T) {
	got := render(chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor)).Root())
	if got != wantScreenshot {
		t.Errorf("screenshot mismatch\ngot  %q\nwant %q", got, wantScreenshot)
	}
}

func TestBgBlueKeepsItsDefaultForeground(t *testing.T) {
	for name, want := range map[string]bool{
		"bgBlack": false,
		"bgBlue":  false,
		"bgRed":   true,
		"bgWhite": true,
		"red":     false,
	} {
		if got := needsBlackText(name); got != want {
			t.Errorf("needsBlackText(%q) = %t, want %t", name, got, want)
		}
	}
}

func TestScreenshotOmitsUnderlineColors(t *testing.T) {
	if strings.Contains(wantScreenshot, "58;5;") {
		t.Error("screenshot must not contain underline color sequences")
	}
}

var errWrite = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errWrite }

func TestRunWritesTheScreenshotForTheDefaultInstance(t *testing.T) {
	var buffer strings.Builder

	if err := run(&buffer); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got, want := buffer.String(), render(chalk.Default.Root()); got != want {
		t.Errorf("run wrote %q, want %q", got, want)
	}
}

func TestRunReportsAWriteFailure(t *testing.T) {
	if err := run(failingWriter{}); !errors.Is(err, errWrite) {
		t.Errorf("run: got %v, want %v", err, errWrite)
	}
}
