// Package ansistyles is a Go port of the `ansi-styles` module vendored into
// chalk v6.0.0 (source/vendor/ansi-styles/index.js).
//
// It carries two features that are not present in upstream `ansi-styles`:
//
//   - Extended underline styles using SGR 4 sub-parameters (4:2, 4:3, 4:4, 4:5).
//   - Underline colors using SGR 58 / 59.
//
// The style table below is transcribed verbatim from the JavaScript source.
// Group order and within-group order are significant: they define the order of
// the exported name slices, which callers iterate.
package ansistyles

import "strconv"

const (
	// ansiBackgroundOffset shifts a foreground SGR code to its background form.
	ansiBackgroundOffset = 10
	// ansiUnderlineOffset shifts a foreground SGR code to its underline-color form.
	ansiUnderlineOffset = 20
)

// escape is the ANSI Control Sequence Introducer prefix.
const escape = "\x1b["

// Group identifies which of the four style groups a style belongs to.
type Group uint8

// The four style groups, in the order they appear in the JavaScript source.
const (
	GroupModifier Group = iota
	GroupColor
	GroupBgColor
	GroupUnderlineColor
)

// Style is a resolved style: the sequence that turns it on and the sequence
// that turns it off.
type Style struct {
	Name  string
	Group Group
	Open  string
	Close string
}

// rawStyle mirrors a `[open, close]` pair from the JavaScript table.
//
// open is a string rather than an int because several styles use multi-part SGR
// parameters ("4:2", "58;5;1") that have no integer form.
type rawStyle struct {
	name  string
	open  string
	close int
}

// The tables below are a direct transcription. Do not reorder.

var rawModifiers = []rawStyle{
	{"reset", "0", 0},
	// 21 isn't widely supported and 22 does the same thing.
	{"bold", "1", 22},
	{"dim", "2", 22},
	{"italic", "3", 23},
	{"underline", "4", 24},
	// Extended underline styles (SGR 4:x sub-parameters). Not in upstream ansi-styles.
	{"underlineDouble", "4:2", 24},
	{"underlineCurly", "4:3", 24},
	{"underlineDotted", "4:4", 24},
	{"underlineDashed", "4:5", 24},
	{"overline", "53", 55},
	{"inverse", "7", 27},
	{"hidden", "8", 28},
	{"strikethrough", "9", 29},
}

var rawColors = []rawStyle{
	{"black", "30", 39},
	{"red", "31", 39},
	{"green", "32", 39},
	{"yellow", "33", 39},
	{"blue", "34", 39},
	{"magenta", "35", 39},
	{"cyan", "36", 39},
	{"white", "37", 39},

	// Bright color.
	{"blackBright", "90", 39},
	{"gray", "90", 39}, // Alias of blackBright.
	{"grey", "90", 39}, // Alias of blackBright.
	{"redBright", "91", 39},
	{"greenBright", "92", 39},
	{"yellowBright", "93", 39},
	{"blueBright", "94", 39},
	{"magentaBright", "95", 39},
	{"cyanBright", "96", 39},
	{"whiteBright", "97", 39},
}

var rawBgColors = []rawStyle{
	{"bgBlack", "40", 49},
	{"bgRed", "41", 49},
	{"bgGreen", "42", 49},
	{"bgYellow", "43", 49},
	{"bgBlue", "44", 49},
	{"bgMagenta", "45", 49},
	{"bgCyan", "46", 49},
	{"bgWhite", "47", 49},

	// Bright color.
	{"bgBlackBright", "100", 49},
	{"bgGray", "100", 49}, // Alias of bgBlackBright.
	{"bgGrey", "100", 49}, // Alias of bgBlackBright.
	{"bgRedBright", "101", 49},
	{"bgGreenBright", "102", 49},
	{"bgYellowBright", "103", 49},
	{"bgBlueBright", "104", 49},
	{"bgMagentaBright", "105", 49},
	{"bgCyanBright", "106", 49},
	{"bgWhiteBright", "107", 49},
}

// Underline color (SGR 58 / 59). Not in upstream ansi-styles.
var rawUnderlineColors = []rawStyle{
	{"underlineBlack", "58;5;0", 59},
	{"underlineRed", "58;5;1", 59},
	{"underlineGreen", "58;5;2", 59},
	{"underlineYellow", "58;5;3", 59},
	{"underlineBlue", "58;5;4", 59},
	{"underlineMagenta", "58;5;5", 59},
	{"underlineCyan", "58;5;6", 59},
	{"underlineWhite", "58;5;7", 59},

	// Bright color.
	{"underlineBlackBright", "58;5;8", 59},
	{"underlineGray", "58;5;8", 59}, // Alias of underlineBlackBright.
	{"underlineGrey", "58;5;8", 59}, // Alias of underlineBlackBright.
	{"underlineRedBright", "58;5;9", 59},
	{"underlineGreenBright", "58;5;10", 59},
	{"underlineYellowBright", "58;5;11", 59},
	{"underlineBlueBright", "58;5;12", 59},
	{"underlineMagentaBright", "58;5;13", 59},
	{"underlineCyanBright", "58;5;14", 59},
	{"underlineWhiteBright", "58;5;15", 59},
}

// Group-level close sequences, used by the parameterized color models.
const (
	// CloseColor ends a foreground color (SGR 39).
	CloseColor = escape + "39m"
	// CloseBgColor ends a background color (SGR 49).
	CloseBgColor = escape + "49m"
	// CloseUnderlineColor ends an underline color (SGR 59).
	CloseUnderlineColor = escape + "59m"
)

// Ordered style slices, one per group.
var (
	// Modifiers lists the modifier styles in source order.
	Modifiers = build(rawModifiers, GroupModifier)
	// Colors lists the foreground color styles in source order.
	Colors = build(rawColors, GroupColor)
	// BgColors lists the background color styles in source order.
	BgColors = build(rawBgColors, GroupBgColor)
	// UnderlineColors lists the underline color styles in source order.
	UnderlineColors = build(rawUnderlineColors, GroupUnderlineColor)

	// All lists every style, groups in source order.
	All = concat(Modifiers, Colors, BgColors, UnderlineColors)

	byName = index(All)

	// ModifierNames mirrors the `modifierNames` export.
	ModifierNames = names(Modifiers)
	// ForegroundColorNames mirrors the `foregroundColorNames` export.
	ForegroundColorNames = names(Colors)
	// BackgroundColorNames mirrors the `backgroundColorNames` export.
	BackgroundColorNames = names(BgColors)
	// UnderlineColorNames mirrors the `underlineColorNames` export.
	UnderlineColorNames = names(UnderlineColors)

	// ColorNames mirrors the `colorNames` export: foreground plus background.
	//
	// Underline colors are deliberately excluded, matching the JavaScript.
	ColorNames = appendAll(ForegroundColorNames, BackgroundColorNames)
)

func build(raw []rawStyle, group Group) []Style {
	out := make([]Style, len(raw))
	for i, r := range raw {
		out[i] = Style{
			Name:  r.name,
			Group: group,
			Open:  escape + r.open + "m",
			Close: escape + strconv.Itoa(r.close) + "m",
		}
	}

	return out
}

func concat(groups ...[]Style) []Style {
	total := 0
	for _, g := range groups {
		total += len(g)
	}

	out := make([]Style, 0, total)
	for _, g := range groups {
		out = append(out, g...)
	}

	return out
}

func index(styles []Style) map[string]Style {
	out := make(map[string]Style, len(styles))
	for _, s := range styles {
		out[s.Name] = s
	}

	return out
}

func names(styles []Style) []string {
	out := make([]string, len(styles))
	for i, s := range styles {
		out[i] = s.Name
	}

	return out
}

func appendAll(lists ...[]string) []string {
	total := 0
	for _, l := range lists {
		total += len(l)
	}

	out := make([]string, 0, total)
	for _, l := range lists {
		out = append(out, l...)
	}

	return out
}

// Lookup returns the style registered under name.
//
// The second result reports whether the name is known.
func Lookup(name string) (Style, bool) {
	s, ok := byName[name]
	return s, ok
}
