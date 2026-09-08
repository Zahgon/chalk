package ansistyles

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// jsRound reproduces ECMAScript's Math.round, which rounds halves toward
// positive infinity.
//
// Go's math.Round rounds halves away from zero, so the two disagree on
// negative halves. Every input reachable here is non-negative, but the
// difference is preserved rather than assumed away.
func jsRound(x float64) float64 {
	return math.Floor(x + 0.5)
}

// RGBToAnsi256 downsamples a 24-bit color to the 256-color palette.
//
// Ported from color-convert via chalk's vendored ansi-styles. The operation
// order is load-bearing for rounding and must not be restructured.
func RGBToAnsi256(red, green, blue uint8) uint8 {
	r, g, b := float64(red), float64(green), float64(blue)

	// The extended greyscale ramp is used for greys, with the exception of
	// black and white: the normal palette only has four greyscale shades.
	if red == green && green == blue {
		if red < 8 {
			return 16
		}

		if red > 248 {
			return 231
		}

		return uint8(jsRound(((r-8)/247)*24) + 232)
	}

	return uint8(16 +
		(36 * jsRound(r/255*5)) +
		(6 * jsRound(g/255*5)) +
		jsRound(b/255*5))
}

// Ansi256ToAnsi downsamples a 256-color palette index to a basic 16-color
// SGR code (30-37 and 90-97).
func Ansi256ToAnsi(code uint8) int {
	if code < 8 {
		return 30 + int(code)
	}

	if code < 16 {
		return 90 + (int(code) - 8)
	}

	var red, green, blue float64

	if code >= 232 {
		red = ((float64(code)-232)*10 + 8) / 255
		green = red
		blue = red
	} else {
		c := int(code) - 16

		remainder := c % 36

		red = math.Floor(float64(c)/36) / 5
		green = math.Floor(float64(remainder)/6) / 5
		blue = float64(remainder%6) / 5
	}

	value := math.Max(math.Max(red, green), blue) * 2

	if value == 0 {
		return 30
	}

	// Bit order is blue<<2 | green<<1 | red, matching the ANSI color layout.
	result := 30 + ((int(jsRound(blue)) << 2) | (int(jsRound(green)) << 1) | int(jsRound(red)))

	// An exact comparison, as in the source. Only a channel of exactly 5/5
	// reaches it; the greyscale ramp tops out below 1.0 and never does.
	if value == 2 {
		result += 60
	}

	return result
}

// hexPattern matches the first run of six or three hex digits anywhere in the
// input.
//
// Go's regexp uses leftmost-first alternation, so the six-digit branch is
// preferred at each position exactly as in the JavaScript original. The
// pattern is deliberately unanchored: "xyz#FF0000!" matches, and "abcd"
// matches only "abc".
var hexPattern = regexp.MustCompile(`(?i)[0-9a-f]{6}|[0-9a-f]{3}`)

// HexToRGB parses a hex color.
//
// It is intentionally lenient, mirroring the JavaScript: the digits may appear
// anywhere in the input, a three-digit form is expanded by doubling each digit,
// and anything unparseable yields black.
func HexToRGB(hex string) (red, green, blue uint8) {
	match := hexPattern.FindString(hex)
	if match == "" {
		return 0, 0, 0
	}

	if len(match) == 3 {
		var sb strings.Builder
		sb.Grow(6)
		for _, c := range match {
			sb.WriteRune(c)
			sb.WriteRune(c)
		}

		match = sb.String()
	}

	// The pattern guarantees six hex digits, so this cannot fail.
	integer, err := strconv.ParseUint(match, 16, 32)
	if err != nil {
		return 0, 0, 0
	}

	return uint8((integer >> 16) & 0xFF), uint8((integer >> 8) & 0xFF), uint8(integer & 0xFF)
}

// HexToAnsi256 converts a hex color to a 256-color palette index.
func HexToAnsi256(hex string) uint8 {
	return RGBToAnsi256(HexToRGB(hex))
}

// RGBToAnsi converts a 24-bit color to a basic 16-color SGR code.
func RGBToAnsi(red, green, blue uint8) int {
	return Ansi256ToAnsi(RGBToAnsi256(red, green, blue))
}

// HexToAnsi converts a hex color to a basic 16-color SGR code.
func HexToAnsi(hex string) int {
	return Ansi256ToAnsi(HexToAnsi256(hex))
}

// Sequence builders. Each corresponds to one of the wrap* closures in the
// JavaScript source.

func wrapAnsi16(offset int, code int) string {
	return escape + strconv.Itoa(code+offset) + "m"
}

func wrapAnsi256(offset int, code uint8) string {
	return escape + strconv.Itoa(38+offset) + ";5;" + strconv.Itoa(int(code)) + "m"
}

func wrapAnsi16m(offset int, red, green, blue uint8) string {
	var sb strings.Builder
	sb.Grow(24)
	sb.WriteString(escape)
	sb.WriteString(strconv.Itoa(38 + offset))
	sb.WriteString(";2;")
	sb.WriteString(strconv.Itoa(int(red)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(green)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(blue)))
	sb.WriteString("m")

	return sb.String()
}

// wrapUnderlineAnsi maps a basic 16-color SGR code to an underline color.
//
// SGR 58 has no basic 16-color form, so the code is remapped to the equivalent
// palette index instead.
func wrapUnderlineAnsi(code int) string {
	index := code - 30
	if code >= 90 {
		index = code - 90 + 8
	}

	return escape + "58;5;" + strconv.Itoa(index) + "m"
}

// Foreground color sequences (SGR 38 family).

// ColorAnsi returns the foreground sequence for a basic 16-color SGR code.
func ColorAnsi(code int) string { return wrapAnsi16(0, code) }

// ColorAnsi256 returns the foreground sequence for a 256-color palette index.
func ColorAnsi256(code uint8) string { return wrapAnsi256(0, code) }

// ColorAnsi16m returns the foreground sequence for a 24-bit color.
func ColorAnsi16m(r, g, b uint8) string { return wrapAnsi16m(0, r, g, b) }

// Background color sequences (SGR 48 family).

// BgColorAnsi returns the background sequence for a basic 16-color SGR code.
func BgColorAnsi(code int) string { return wrapAnsi16(ansiBackgroundOffset, code) }

// BgColorAnsi256 returns the background sequence for a 256-color palette index.
func BgColorAnsi256(code uint8) string { return wrapAnsi256(ansiBackgroundOffset, code) }

// BgColorAnsi16m returns the background sequence for a 24-bit color.
func BgColorAnsi16m(r, g, b uint8) string { return wrapAnsi16m(ansiBackgroundOffset, r, g, b) }

// Underline color sequences (SGR 58 family).

// UnderlineColorAnsi returns the underline sequence for a basic 16-color SGR code.
func UnderlineColorAnsi(code int) string { return wrapUnderlineAnsi(code) }

// UnderlineColorAnsi256 returns the underline sequence for a 256-color palette index.
func UnderlineColorAnsi256(code uint8) string { return wrapAnsi256(ansiUnderlineOffset, code) }

// UnderlineColorAnsi16m returns the underline sequence for a 24-bit color.
func UnderlineColorAnsi16m(r, g, b uint8) string { return wrapAnsi16m(ansiUnderlineOffset, r, g, b) }
