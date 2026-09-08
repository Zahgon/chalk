// Package chalk renders styled terminal text using ANSI escape sequences.
//
// It is a port of chalk v6.0.0 for Node.js and reproduces its rendering
// behaviour exactly, including terminal capability detection, automatic
// downsampling of 24-bit colors, correct handling of nested styles, and
// per-line reopening of styles across line breaks.
//
// The zero value of [Style] is usable and renders through [Default].
//
// Basic use:
//
//	fmt.Println(chalk.Red().Sprint("hello"))
//	fmt.Println(chalk.Blue().Bold().Underline().Sprint("world"))
//	fmt.Println(chalk.RGB(255, 136, 0).Sprint("orange"))
//
// A [Style] is an immutable value. Each method returns a new [Style], so chains
// can be stored and reused safely:
//
//	warn := chalk.Yellow().Bold()
//	fmt.Println(warn.Sprint("careful"))
package chalk

//go:generate go run ./gen -o styles_gen.go

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/chalk/chalk-go/supportscolor"
)

// ErrInvalidLevel is returned when a color level outside 0-3 is requested.
//
// Test with [errors.Is].
var ErrInvalidLevel = errors.New("the level should be an integer from 0 to 3")

// Color level constants.
//
// A level is the ceiling on what a [Style] may emit; colors above it are
// downsampled automatically.
const (
	// LevelNone disables all color output.
	LevelNone = 0
	// LevelBasic enables the 16 basic colors.
	LevelBasic = 1
	// LevelAnsi256 enables the 256-color palette.
	LevelAnsi256 = 2
	// LevelTrueColor enables 24-bit color.
	LevelTrueColor = 3
)

// Instance owns a color level. Styles derived from it read that level when
// they render, so changing the level affects existing chains.
//
// An Instance is safe for concurrent use.
type Instance struct {
	level atomic.Int32
}

// Package-level instances matching chalk's `chalk` and `chalkStderr` exports.
//
// Their levels are detected from the environment at process start.
var (
	// Default renders according to standard output's detected capabilities.
	Default = newDetected(supportscolor.Stdout())
	// Stderr renders according to standard error's detected capabilities.
	Stderr = newDetected(supportscolor.Stderr())
)

func newDetected(support *supportscolor.ColorSupport) *Instance {
	instance := &Instance{}
	if support != nil {
		instance.level.Store(int32(support.Level))
	}

	return instance
}

// An Option configures a new [Instance].
type Option func(*options) error

type options struct {
	// A pointer, so that "not supplied" stays distinct from "level 0".
	level *int
}

// WithLevel fixes the color level instead of detecting it.
//
// It returns an error wrapping [ErrInvalidLevel] if level is outside 0-3.
func WithLevel(level int) Option {
	return func(o *options) error {
		if err := validateLevel(level); err != nil {
			return err
		}

		o.level = &level

		return nil
	}
}

func validateLevel(level int) error {
	if level < LevelNone || level > LevelTrueColor {
		return fmt.Errorf("chalk: %d: %w", level, ErrInvalidLevel)
	}

	return nil
}

// New returns an Instance with its own color level.
//
// Without [WithLevel] the level is detected from standard output, matching
// [Default].
func New(opts ...Option) (*Instance, error) {
	var o options

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	if o.level == nil {
		return newDetected(supportscolor.Stdout()), nil
	}

	instance := &Instance{}
	instance.level.Store(int32(*o.level))

	return instance, nil
}

// MustNew is like [New] but panics if an option is invalid.
//
// It is intended for package-level initialisation with fixed arguments.
func MustNew(opts ...Option) *Instance {
	instance, err := New(opts...)
	if err != nil {
		panic(err)
	}

	return instance
}

// Level reports the current color level.
func (i *Instance) Level() int {
	if i == nil {
		return LevelNone
	}

	return int(i.level.Load())
}

// SetLevel changes the color level.
//
// Styles already derived from this Instance observe the change. It returns an
// error wrapping [ErrInvalidLevel] if level is outside 0-3.
func (i *Instance) SetLevel(level int) error {
	if err := validateLevel(level); err != nil {
		return err
	}

	i.level.Store(int32(level))

	return nil
}

// Style is an immutable, chainable set of styles bound to an [Instance].
//
// The zero value renders through [Default] and applies no styling.
type Style struct {
	instance *Instance
	styler   *styler
	isEmpty  bool
}

// Root returns the unstyled Style bound to this Instance.
//
// It is the starting point for chains built from a custom Instance.
func (i *Instance) Root() Style {
	return Style{instance: i}
}

// owner returns the bound Instance, defaulting to [Default] for the zero value.
func (s Style) owner() *Instance {
	if s.instance == nil {
		return Default
	}

	return s.instance
}

// Instance returns the [Instance] this Style renders through.
//
// Use it to change the level of a chain, which chalk spells `chain.level = n`:
//
//	chain.Instance().SetLevel(chalk.LevelNone)
func (s Style) Instance() *Instance {
	return s.owner()
}

// Level reports the color level this Style renders at.
func (s Style) Level() int {
	return s.owner().Level()
}

// with returns a copy of s with one more style pushed onto its chain.
func (s Style) with(open, close string) Style {
	return Style{
		instance: s.instance,
		styler:   newStyler(open, close, s.styler),
		isEmpty:  s.isEmpty,
	}
}

// Visible returns a Style that renders only when colors are enabled.
//
// At [LevelNone] it produces an empty string rather than the plain text.
func (s Style) Visible() Style {
	return Style{instance: s.instance, styler: s.styler, isEmpty: true}
}

// By returns the style registered under an ANSI style name, such as "red",
// "bgBlue", "underlineCurly", or "underlineRedBright".
//
// It reports an error for unknown names. Use it when the style is chosen at
// run time; prefer the generated methods otherwise.
func (s Style) By(name string) (Style, error) {
	style, ok := lookupStyle(name)
	if !ok {
		return Style{}, fmt.Errorf("chalk: unknown style %q", name)
	}

	return s.with(style.Open, style.Close), nil
}

// Sprint renders its operands as styled text.
//
// Operands are formatted with [fmt.Sprint] rules and joined with a single
// space, matching chalk's `chalk.red('a', 'b')`. This differs from
// [fmt.Sprint], which only inserts spaces between non-string operands.
func (s Style) Sprint(a ...any) string {
	return s.render(joinOperands(a))
}

// Sprintf renders text formatted according to a format specifier.
func (s Style) Sprintf(format string, a ...any) string {
	return s.render(fmt.Sprintf(format, a...))
}

// Sprintln renders its operands joined by a single space, with a trailing
// newline appended after the closing escape sequence.
func (s Style) Sprintln(a ...any) string {
	return s.render(joinOperands(a)) + "\n"
}

// Print writes styled text to standard output.
func (s Style) Print(a ...any) (int, error) {
	return fmt.Fprint(os.Stdout, s.Sprint(a...))
}

// Printf writes formatted styled text to standard output.
func (s Style) Printf(format string, a ...any) (int, error) {
	return fmt.Fprint(os.Stdout, s.Sprintf(format, a...))
}

// Println writes styled text to standard output, followed by a newline.
func (s Style) Println(a ...any) (int, error) {
	return fmt.Fprint(os.Stdout, s.Sprintln(a...))
}

// Fprint writes styled text to w.
func (s Style) Fprint(w io.Writer, a ...any) (int, error) {
	return fmt.Fprint(w, s.Sprint(a...))
}

// Fprintf writes formatted styled text to w.
func (s Style) Fprintf(w io.Writer, format string, a ...any) (int, error) {
	return fmt.Fprint(w, s.Sprintf(format, a...))
}

// Fprintln writes styled text to w, followed by a newline.
func (s Style) Fprintln(w io.Writer, a ...any) (int, error) {
	return fmt.Fprint(w, s.Sprintln(a...))
}

// joinOperands formats operands and joins them with a single space.
func joinOperands(a []any) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(a[0])
	case 2:
		return fmt.Sprint(a[0]) + " " + fmt.Sprint(a[1])
	default:
		parts := make([]string, len(a))
		for i, v := range a {
			parts[i] = fmt.Sprint(v)
		}

		return strings.Join(parts, " ")
	}
}

// render wraps text in this Style's escape sequences.
//
// Ported from applyStyle in source/index.js.
func (s Style) render(text string) string {
	if s.owner().Level() <= LevelNone || text == "" {
		if s.isEmpty {
			return ""
		}

		return text
	}

	current := s.styler
	if current == nil {
		return text
	}

	openAll, closeAll := current.openAll, current.closeAll

	// Nested styles: an inner style's close sequence would end this style too,
	// so reopen this style after each one. Walking up the chain handles a
	// nested style that closed several levels at once.
	if strings.IndexByte(text, '\x1b') != -1 {
		for node := current; node != nil; node = node.parent {
			text = stringReplaceAll(text, node.close, node.open)
		}
	}

	// Styles must not bleed across line breaks: terminals do not reliably
	// carry a background color to the start of the next line, and a styled
	// region spanning a break confuses pagers.
	if lfIndex := strings.IndexByte(text, '\n'); lfIndex != -1 {
		text = stringEncaseCRLFWithFirstIndex(text, closeAll, openAll, lfIndex)
	}

	return openAll + text + closeAll
}
