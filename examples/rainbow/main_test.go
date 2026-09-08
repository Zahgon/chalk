package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/chalk/chalk-go"
)

// hslToHex expectations captured from `color-convert`'s hsl.hex in Node.
func TestHSLToHexMatchesColorConvert(t *testing.T) {
	cases := []struct {
		hue, saturation, lightness float64
		want                       string
	}{
		{0, 100, 50, "FF0000"},
		{1, 100, 50, "FF0400"},
		{30, 100, 50, "FF8000"},
		{59, 100, 50, "FFFB00"},
		{60, 100, 50, "FFFF00"},
		{90, 100, 50, "80FF00"},
		{119, 100, 50, "04FF00"},
		{120, 100, 50, "00FF00"},
		{180, 100, 50, "00FFFF"},
		{210, 100, 50, "007FFF"},
		{239, 100, 50, "0004FF"},
		{240, 100, 50, "0000FF"},
		{300, 100, 50, "FF00FF"},
		{359, 100, 50, "FF0004"},
		{200, 0, 50, "808080"},
		{200, 50, 25, "204A60"},
		{200, 50, 75, "9FCADF"},
		{13.846153846153847, 100, 50, "FF3B00"},
	}

	for _, tc := range cases {
		got := hslToHex(tc.hue, tc.saturation, tc.lightness)
		if got != tc.want {
			t.Errorf("hslToHex(%v, %v, %v) = %q, want %q", tc.hue, tc.saturation, tc.lightness, got, tc.want)
		}
	}
}

// Expectations captured from examples/rainbow.js running against the original
// chalk source at level 3.
func TestRainbowMatchesJavaScript(t *testing.T) {
	const banner = "We hope you enjoy Chalk! <3"
	style := chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor)).Root()

	cases := map[int]string{
		0:    "\x1b[38;2;255;0;0mW\x1b[39m\x1b[38;2;255;70;0me\x1b[39m \x1b[38;2;255;139;0mh\x1b[39m\x1b[38;2;255;209;0mo\x1b[39m\x1b[38;2;232;255;0mp\x1b[39m\x1b[38;2;162;255;0me\x1b[39m \x1b[38;2;93;255;0my\x1b[39m\x1b[38;2;23;255;0mo\x1b[39m\x1b[38;2;0;255;46mu\x1b[39m \x1b[38;2;0;255;116me\x1b[39m\x1b[38;2;0;255;185mn\x1b[39m\x1b[38;2;0;255;255mj\x1b[39m\x1b[38;2;0;185;255mo\x1b[39m\x1b[38;2;0;116;255my\x1b[39m \x1b[38;2;0;46;255mC\x1b[39m\x1b[38;2;23;0;255mh\x1b[39m\x1b[38;2;93;0;255ma\x1b[39m\x1b[38;2;162;0;255ml\x1b[39m\x1b[38;2;232;0;255mk\x1b[39m\x1b[38;2;255;0;209m!\x1b[39m \x1b[38;2;255;0;139m<\x1b[39m\x1b[38;2;255;0;70m3\x1b[39m",
		1:    "\x1b[38;2;255;4;0mW\x1b[39m\x1b[38;2;255;74;0me\x1b[39m \x1b[38;2;255;143;0mh\x1b[39m\x1b[38;2;255;213;0mo\x1b[39m\x1b[38;2;228;255;0mp\x1b[39m\x1b[38;2;158;255;0me\x1b[39m \x1b[38;2;88;255;0my\x1b[39m\x1b[38;2;19;255;0mo\x1b[39m\x1b[38;2;0;255;51mu\x1b[39m \x1b[38;2;0;255;120me\x1b[39m\x1b[38;2;0;255;190mn\x1b[39m\x1b[38;2;0;251;255mj\x1b[39m\x1b[38;2;0;181;255mo\x1b[39m\x1b[38;2;0;112;255my\x1b[39m \x1b[38;2;0;42;255mC\x1b[39m\x1b[38;2;27;0;255mh\x1b[39m\x1b[38;2;97;0;255ma\x1b[39m\x1b[38;2;167;0;255ml\x1b[39m\x1b[38;2;236;0;255mk\x1b[39m\x1b[38;2;255;0;204m!\x1b[39m \x1b[38;2;255;0;135m<\x1b[39m\x1b[38;2;255;0;65m3\x1b[39m",
		7:    "\x1b[38;2;255;30;0mW\x1b[39m\x1b[38;2;255;99;0me\x1b[39m \x1b[38;2;255;169;0mh\x1b[39m\x1b[38;2;255;238;0mo\x1b[39m\x1b[38;2;202;255;0mp\x1b[39m\x1b[38;2;133;255;0me\x1b[39m \x1b[38;2;63;255;0my\x1b[39m\x1b[38;2;0;255;7mo\x1b[39m\x1b[38;2;0;255;76mu\x1b[39m \x1b[38;2;0;255;146me\x1b[39m\x1b[38;2;0;255;215mn\x1b[39m\x1b[38;2;0;225;255mj\x1b[39m\x1b[38;2;0;156;255mo\x1b[39m\x1b[38;2;0;86;255my\x1b[39m \x1b[38;2;0;17;255mC\x1b[39m\x1b[38;2;53;0;255mh\x1b[39m\x1b[38;2;122;0;255ma\x1b[39m\x1b[38;2;192;0;255ml\x1b[39m\x1b[38;2;255;0;248mk\x1b[39m\x1b[38;2;255;0;179m!\x1b[39m \x1b[38;2;255;0;109m<\x1b[39m\x1b[38;2;255;0;40m3\x1b[39m",
		123:  "\x1b[38;2;0;255;13mW\x1b[39m\x1b[38;2;0;255;82me\x1b[39m \x1b[38;2;0;255;152mh\x1b[39m\x1b[38;2;0;255;221mo\x1b[39m\x1b[38;2;0;219;255mp\x1b[39m\x1b[38;2;0;150;255me\x1b[39m \x1b[38;2;0;80;255my\x1b[39m\x1b[38;2;0;10;255mo\x1b[39m\x1b[38;2;59;0;255mu\x1b[39m \x1b[38;2;129;0;255me\x1b[39m\x1b[38;2;198;0;255mn\x1b[39m\x1b[38;2;255;0;242mj\x1b[39m\x1b[38;2;255;0;173mo\x1b[39m\x1b[38;2;255;0;103my\x1b[39m \x1b[38;2;255;0;34mC\x1b[39m\x1b[38;2;255;36;0mh\x1b[39m\x1b[38;2;255;105;0ma\x1b[39m\x1b[38;2;255;175;0ml\x1b[39m\x1b[38;2;255;245;0mk\x1b[39m\x1b[38;2;196;255;0m!\x1b[39m \x1b[38;2;126;255;0m<\x1b[39m\x1b[38;2;57;255;0m3\x1b[39m",
		359:  "\x1b[38;2;255;0;4mW\x1b[39m\x1b[38;2;255;65;0me\x1b[39m \x1b[38;2;255;135;0mh\x1b[39m\x1b[38;2;255;204;0mo\x1b[39m\x1b[38;2;236;255;0mp\x1b[39m\x1b[38;2;167;255;0me\x1b[39m \x1b[38;2;97;255;0my\x1b[39m\x1b[38;2;27;255;0mo\x1b[39m\x1b[38;2;0;255;42mu\x1b[39m \x1b[38;2;0;255;112me\x1b[39m\x1b[38;2;0;255;181mn\x1b[39m\x1b[38;2;0;255;251mj\x1b[39m\x1b[38;2;0;190;255mo\x1b[39m\x1b[38;2;0;120;255my\x1b[39m \x1b[38;2;0;51;255mC\x1b[39m\x1b[38;2;19;0;255mh\x1b[39m\x1b[38;2;88;0;255ma\x1b[39m\x1b[38;2;158;0;255ml\x1b[39m\x1b[38;2;228;0;255mk\x1b[39m\x1b[38;2;255;0;213m!\x1b[39m \x1b[38;2;255;0;143m<\x1b[39m\x1b[38;2;255;0;74m3\x1b[39m",
		1799: "\x1b[38;2;255;0;4mW\x1b[39m\x1b[38;2;255;65;0me\x1b[39m \x1b[38;2;255;135;0mh\x1b[39m\x1b[38;2;255;204;0mo\x1b[39m\x1b[38;2;236;255;0mp\x1b[39m\x1b[38;2;167;255;0me\x1b[39m \x1b[38;2;97;255;0my\x1b[39m\x1b[38;2;27;255;0mo\x1b[39m\x1b[38;2;0;255;42mu\x1b[39m \x1b[38;2;0;255;112me\x1b[39m\x1b[38;2;0;255;181mn\x1b[39m\x1b[38;2;0;255;251mj\x1b[39m\x1b[38;2;0;190;255mo\x1b[39m\x1b[38;2;0;120;255my\x1b[39m \x1b[38;2;0;51;255mC\x1b[39m\x1b[38;2;19;0;255mh\x1b[39m\x1b[38;2;88;0;255ma\x1b[39m\x1b[38;2;158;0;255ml\x1b[39m\x1b[38;2;228;0;255mk\x1b[39m\x1b[38;2;255;0;213m!\x1b[39m \x1b[38;2;255;0;143m<\x1b[39m\x1b[38;2;255;0;74m3\x1b[39m",
	}

	for offset, want := range cases {
		if got := rainbow(style, banner, offset); got != want {
			t.Errorf("rainbow(offset=%d) mismatch\ngot  %q\nwant %q", offset, got, want)
		}
	}

	if got := rainbow(style, "", 5); got != "" {
		t.Errorf("rainbow of empty string = %q, want empty", got)
	}

	const wantSpaces = "\x1b[38;2;128;255;0ma\x1b[39m \x1b[38;2;127;0;255mb\x1b[39m"
	if got := rainbow(style, "a b", 90); got != wantSpaces {
		t.Errorf("rainbow spaces mismatch\ngot  %q\nwant %q", got, wantSpaces)
	}
}

func TestRainbowIsPlainAtLevelZero(t *testing.T) {
	style := chalk.MustNew(chalk.WithLevel(chalk.LevelNone)).Root()
	if got := rainbow(style, "abc", 0); got != "abc" {
		t.Errorf("rainbow at level 0 = %q, want %q", got, "abc")
	}
}

func TestUpdaterErasesPreviousFrame(t *testing.T) {
	var buffer strings.Builder
	updater := &lineUpdater{writer: &buffer}

	if err := updater.update("first"); err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != "first\n" {
		t.Errorf("first frame = %q, want %q", got, "first\n")
	}

	buffer.Reset()
	if err := updater.update("second"); err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != eraseLine+cursorLeft+"second\n" {
		t.Errorf("second frame = %q", got)
	}
}

var errWrite = errors.New("write failed")

type failAfter struct{ remaining int }

func (w *failAfter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errWrite
	}

	w.remaining--

	return len(p), nil
}

func TestAnimateHidesTheCursorForTheWholeRun(t *testing.T) {
	var buffer strings.Builder

	if err := animate(&buffer, "hi", 3, 0); err != nil {
		t.Fatalf("animate: %v", err)
	}

	got := buffer.String()

	if !strings.HasPrefix(got, hideCursor) {
		t.Errorf("animate must hide the cursor before the first frame, got %q", got)
	}

	if !strings.HasSuffix(got, showCursor) {
		t.Errorf("animate must restore the cursor when it returns, got %q", got)
	}

	if n := strings.Count(got, "\n"); n != 3 {
		t.Errorf("animate wrote %d frames, want 3", n)
	}
}

func TestAnimateStopsAtTheFirstWriteError(t *testing.T) {
	if err := animate(&failAfter{}, "hi", 1, 0); !errors.Is(err, errWrite) {
		t.Errorf("failing to hide the cursor: got %v, want %v", err, errWrite)
	}

	if err := animate(&failAfter{remaining: 1}, "hi", 2, 0); !errors.Is(err, errWrite) {
		t.Errorf("failing to write a frame: got %v, want %v", err, errWrite)
	}
}
