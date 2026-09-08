// Command screenshot prints one sample of every style, for use as a readme
// screenshot. It is a port of examples/screenshot.js.
package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/chalk/chalk-go"
	"github.com/chalk/chalk-go/ansistyles"
)

// skipped mirrors the exclusion list in examples/screenshot.js. `overline` is
// skipped because almost no terminal supports it, so it cannot be shown off.
var skipped = map[string]bool{
	"reset":    true,
	"hidden":   true,
	"grey":     true,
	"bgGray":   true,
	"bgGrey":   true,
	"overline": true,
}

// showcased reports whether a style appears in the screenshot.
//
// The JavaScript example iterates the keys of the upstream `ansi-styles`
// package rather than the copy vendored into chalk, so the underline color
// group and the `underlineDouble`-style modifiers are absent there. Excluding
// them here keeps the two screenshots identical.
func showcased(style ansistyles.Style) bool {
	if style.Group == ansistyles.GroupUnderlineColor {
		return false
	}
	if strings.HasPrefix(style.Name, "underline") && style.Name != "underline" {
		return false
	}
	return !skipped[style.Name] && !strings.HasSuffix(style.Name, "Bright")
}

// needsBlackText reports whether a background style is light enough to need
// black text. It ports the /^bg[^B]/v test, which deliberately also excludes
// `bgBlue` because its third character is an uppercase B.
func needsBlackText(name string) bool {
	return strings.HasPrefix(name, "bg") && len(name) > 2 && name[2] != 'B'
}

func render(root chalk.Style) string {
	var out strings.Builder
	for _, style := range ansistyles.All {
		if !showcased(style) {
			continue
		}

		text := style.Name
		if needsBlackText(style.Name) {
			text = root.Black().Sprint(text)
		}

		styled, err := root.By(style.Name)
		if err != nil {
			panic(err)
		}

		out.WriteString(styled.Sprint(text))
		out.WriteString(" ")
	}
	return out.String()
}

func run(w io.Writer) error {
	buffered := bufio.NewWriter(w)
	if _, err := io.WriteString(buffered, render(chalk.Default.Root())); err != nil {
		return err
	}
	return buffered.Flush()
}

func main() {
	if err := run(os.Stdout); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
