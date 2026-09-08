// Command rainbow animates a rainbow-colored string in place. It is a port of
// examples/rainbow.js, with the `color-convert` and `log-update` dependencies
// reimplemented locally.
package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/chalk/chalk-go"
)

// isIgnored ports /[^\u{0021}-\u{007E}]/, which selects every character
// outside printable ASCII. Ignored characters are emitted unstyled and do not
// consume a hue step.
func isIgnored(r rune) bool {
	return r < 0x21 || r > 0x7E
}

// hslToRGB ports color-convert's hsl.rgb. Values are returned unrounded so
// that rounding happens exactly once, in hslToHex, as it does in JavaScript.
func hslToRGB(hue, saturation, lightness float64) (r, g, b float64) {
	h := hue / 360
	s := saturation / 100
	l := lightness / 100

	if s == 0 {
		value := l * 255
		return value, value, value
	}

	var t2 float64
	if l < 0.5 {
		t2 = l * (1 + s)
	} else {
		t2 = l + s - l*s
	}
	t1 := 2*l - t2

	channel := func(index float64) float64 {
		t3 := h + 1.0/3.0*-(index-1)
		if t3 < 0 {
			t3++
		}
		if t3 > 1 {
			t3--
		}

		switch {
		case 6*t3 < 1:
			return (t1 + (t2-t1)*6*t3) * 255
		case 2*t3 < 1:
			return t2 * 255
		case 3*t3 < 2:
			return (t1 + (t2-t1)*(2.0/3.0-t3)*6) * 255
		default:
			return t1 * 255
		}
	}

	return channel(0), channel(1), channel(2)
}

// jsRound reproduces JavaScript's Math.round, which rounds halves toward
// positive infinity rather than away from zero like Go's math.Round.
func jsRound(value float64) float64 {
	return math.Floor(value + 0.5)
}

func hslToHex(hue, saturation, lightness float64) string {
	r, g, b := hslToRGB(hue, saturation, lightness)
	return fmt.Sprintf("%02X%02X%02X",
		uint8(int64(jsRound(r))&0xFF),
		uint8(int64(jsRound(g))&0xFF),
		uint8(int64(jsRound(b))&0xFF),
	)
}

func rainbow(style chalk.Style, text string, offset int) string {
	if text == "" {
		return text
	}

	printable := 0
	for _, r := range text {
		if !isIgnored(r) {
			printable++
		}
	}
	if printable == 0 {
		return text
	}

	hueStep := 360 / float64(printable)
	hue := math.Mod(float64(offset), 360)

	var out strings.Builder
	for _, r := range text {
		if isIgnored(r) {
			out.WriteRune(r)
			continue
		}

		out.WriteString(style.Hex(hslToHex(hue, 100, 50)).Sprint(string(r)))
		hue = math.Mod(hue+hueStep, 360)
	}
	return out.String()
}

// lineUpdater rewrites the same block of lines on every frame, replacing the
// `log-update` dependency.
type lineUpdater struct {
	writer io.Writer
	lines  int
}

const (
	eraseLine  = "\x1b[2K"
	cursorUp   = "\x1b[1A"
	cursorLeft = "\x1b[G"
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
)

func (u *lineUpdater) update(text string) error {
	var frame strings.Builder
	for i := 0; i < u.lines; i++ {
		frame.WriteString(eraseLine)
		if i < u.lines-1 {
			frame.WriteString(cursorUp)
		}
	}
	if u.lines > 0 {
		frame.WriteString(cursorLeft)
	}

	frame.WriteString(text)
	frame.WriteString("\n")
	u.lines = strings.Count(text, "\n") + 1

	_, err := io.WriteString(u.writer, frame.String())
	return err
}

func animate(w io.Writer, text string, frames int, delay time.Duration) error {
	updater := &lineUpdater{writer: w}
	style := chalk.Default.Root()

	if _, err := io.WriteString(w, hideCursor); err != nil {
		return err
	}
	defer io.WriteString(w, showCursor)

	for frame := 0; frame < frames; frame++ {
		if err := updater.update(rainbow(style, text, frame)); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}

func main() {
	frames := flag.Int("frames", 360*5, "number of animation frames to render")
	text := flag.String("text", "We hope you enjoy Chalk! <3", "text to animate")
	flag.Parse()

	fmt.Println()
	if err := animate(os.Stdout, *text, *frames, 2*time.Millisecond); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println()
}
