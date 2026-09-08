package chalk

import "github.com/chalk/chalk-go/ansistyles"

// colorTarget bundles the three sequence builders for one color slot
// (foreground, background, or underline) with that slot's close sequence.
type colorTarget struct {
	ansi    func(code int) string
	ansi256 func(code uint8) string
	ansi16m func(r, g, b uint8) string
	close   string
}

var (
	fgTarget = colorTarget{
		ansi:    ansistyles.ColorAnsi,
		ansi256: ansistyles.ColorAnsi256,
		ansi16m: ansistyles.ColorAnsi16m,
		close:   ansistyles.CloseColor,
	}

	bgTarget = colorTarget{
		ansi:    ansistyles.BgColorAnsi,
		ansi256: ansistyles.BgColorAnsi256,
		ansi16m: ansistyles.BgColorAnsi16m,
		close:   ansistyles.CloseBgColor,
	}

	underlineTarget = colorTarget{
		ansi:    ansistyles.UnderlineColorAnsi,
		ansi256: ansistyles.UnderlineColorAnsi256,
		ansi16m: ansistyles.UnderlineColorAnsi16m,
		close:   ansistyles.CloseUnderlineColor,
	}
)

// The three color models, each downsampling to whatever the level allows.
//
// Level 0 still produces a sequence. It is never emitted, because rendering
// short-circuits before writing any escapes, but resolving it unconditionally
// keeps these total.

func (t colorTarget) rgb(level int, r, g, b uint8) string {
	switch {
	case level >= LevelTrueColor:
		return t.ansi16m(r, g, b)
	case level == LevelAnsi256:
		return t.ansi256(ansistyles.RGBToAnsi256(r, g, b))
	default:
		return t.ansi(ansistyles.RGBToAnsi(r, g, b))
	}
}

func (t colorTarget) hex(level int, hex string) string {
	switch {
	case level >= LevelTrueColor:
		return t.ansi16m(ansistyles.HexToRGB(hex))
	case level == LevelAnsi256:
		return t.ansi256(ansistyles.HexToAnsi256(hex))
	default:
		return t.ansi(ansistyles.HexToAnsi(hex))
	}
}

func (t colorTarget) ansi256Code(level int, code uint8) string {
	if level >= LevelAnsi256 {
		return t.ansi256(code)
	}

	return t.ansi(ansistyles.Ansi256ToAnsi(code))
}

// The color level is read here, when the color is chosen, and the resulting
// sequence is fixed in the returned Style. Lowering the level afterwards
// disables output but does not re-downsample an already chosen color.

// RGB styles text with a 24-bit foreground color, downsampled if necessary.
func (s Style) RGB(r, g, b uint8) Style {
	return s.with(fgTarget.rgb(s.Level(), r, g, b), fgTarget.close)
}

// BgRGB styles text with a 24-bit background color, downsampled if necessary.
func (s Style) BgRGB(r, g, b uint8) Style {
	return s.with(bgTarget.rgb(s.Level(), r, g, b), bgTarget.close)
}

// UnderlineRGB styles text with a 24-bit underline color, downsampled if
// necessary.
func (s Style) UnderlineRGB(r, g, b uint8) Style {
	return s.with(underlineTarget.rgb(s.Level(), r, g, b), underlineTarget.close)
}

// Hex styles text with a foreground color given as hex, such as "#ff8800".
//
// Parsing is lenient: the first run of six or three hex digits anywhere in the
// string is used, a three-digit form is expanded by doubling each digit, and an
// unparseable string yields black.
func (s Style) Hex(hex string) Style {
	return s.with(fgTarget.hex(s.Level(), hex), fgTarget.close)
}

// BgHex styles text with a background color given as hex.
//
// It parses hex as leniently as [Style.Hex].
func (s Style) BgHex(hex string) Style {
	return s.with(bgTarget.hex(s.Level(), hex), bgTarget.close)
}

// UnderlineHex styles text with an underline color given as hex.
//
// It parses hex as leniently as [Style.Hex].
func (s Style) UnderlineHex(hex string) Style {
	return s.with(underlineTarget.hex(s.Level(), hex), underlineTarget.close)
}

// Ansi256 styles text with a foreground color from the 256-color palette.
func (s Style) Ansi256(code uint8) Style {
	return s.with(fgTarget.ansi256Code(s.Level(), code), fgTarget.close)
}

// BgAnsi256 styles text with a background color from the 256-color palette.
func (s Style) BgAnsi256(code uint8) Style {
	return s.with(bgTarget.ansi256Code(s.Level(), code), bgTarget.close)
}

// UnderlineAnsi256 styles text with an underline color from the 256-color
// palette.
func (s Style) UnderlineAnsi256(code uint8) Style {
	return s.with(underlineTarget.ansi256Code(s.Level(), code), underlineTarget.close)
}
