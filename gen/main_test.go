package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/chalk/chalk-go/ansistyles"
)

// styles_gen.go is committed, so a broken generator would not be caught by
// compiling the library. These tests check the generator itself: the naming
// and wording rules it applies, the shape of what it emits, and that the
// committed file is what it produces today.

func TestExportedUppercasesTheFirstLetter(t *testing.T) {
	cases := map[string]string{
		"red":                  "Red",
		"bgRedBright":          "BgRedBright",
		"underlineBlackBright": "UnderlineBlackBright",
	}

	for name, want := range cases {
		if got := exported(name); got != want {
			t.Errorf("exported(%q): got %q, want %q", name, got, want)
		}
	}

	// Every table entry must survive the rule as a valid exported Go
	// identifier, or the generated file would not compile.
	for _, style := range ansistyles.All {
		got := exported(style.Name)

		if got == style.Name {
			t.Errorf("exported(%q) is unchanged, so it would not be exported", style.Name)
		}

		if !strings.EqualFold(got, style.Name) {
			t.Errorf("exported(%q): got %q, which differs by more than case", style.Name, got)
		}
	}
}

func TestColorPhraseReadsAsEnglish(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		want   string
	}{
		{"red", "", "red"},
		{"blackBright", "", "bright black"},
		{"gray", "", "gray"},
		{"grey", "", "grey"},
		{"bgRed", "bg", "red"},
		{"bgRedBright", "bg", "bright red"},
		{"bgGray", "bg", "gray"},
		{"underlineWhite", "underline", "white"},
		{"underlineWhiteBright", "underline", "bright white"},
		{"underlineGrey", "underline", "grey"},
	}

	for _, c := range cases {
		if got := colorPhrase(c.name, c.prefix); got != c.want {
			t.Errorf("colorPhrase(%q, %q): got %q, want %q", c.name, c.prefix, got, c.want)
		}
	}
}

func TestSummaryDescribesEveryTableEntry(t *testing.T) {
	// Compared against the table itself rather than a copy of the wording,
	// so this test cannot become the place the sentences are really defined.
	for name, text := range descriptions {
		style, ok := ansistyles.Lookup(name)
		if !ok {
			t.Errorf("descriptions names %q, which is not in the style table", name)
			continue
		}

		if got := summary(style); got != text {
			t.Errorf("summary(%q): got %q, want the hand-written %q", name, got, text)
		}
	}

	// A color's sentence is a group lead-in followed by the color phrase.
	// The lead-in is read back out of the result instead of being spelled
	// here, then held to being shared by its group and by no other.
	prefixes := map[ansistyles.Group]string{
		ansistyles.GroupColor:          "",
		ansistyles.GroupBgColor:        "bg",
		ansistyles.GroupUnderlineColor: "underline",
	}
	leadIns := map[ansistyles.Group]string{}

	for _, style := range ansistyles.All {
		prefix, isColor := prefixes[style.Group]
		if !isColor {
			continue
		}

		text, phrase := summary(style), colorPhrase(style.Name, prefix)
		if !strings.HasSuffix(text, phrase) {
			t.Errorf("summary(%q): got %q, want it to end in the color phrase %q", style.Name, text, phrase)
			continue
		}

		head := strings.TrimSuffix(text, phrase)
		if seen, ok := leadIns[style.Group]; ok && seen != head {
			t.Errorf("group %d has two lead-ins: %q and %q", style.Group, seen, head)
		}

		leadIns[style.Group] = head
	}

	if len(leadIns) != len(prefixes) {
		t.Errorf("described %d color groups, want %d", len(leadIns), len(prefixes))
	}

	owners := map[string]ansistyles.Group{}
	for group, head := range leadIns {
		if other, ok := owners[head]; ok {
			t.Errorf("groups %d and %d share the lead-in %q", other, group, head)
		}

		owners[head] = group
	}

	// A modifier with no hand-written description falls back to the style
	// name. No table entry does today, so it is checked with synthetic
	// styles rather than left untested.
	blink := summary(ansistyles.Style{Name: "blink", Group: ansistyles.GroupModifier})
	if !strings.Contains(blink, "blink") {
		t.Errorf("summary for an undescribed modifier: got %q, want it to name the style", blink)
	}

	if flash := summary(ansistyles.Style{Name: "flash", Group: ansistyles.GroupModifier}); flash == blink {
		t.Errorf("summary for an undescribed modifier does not vary with the name: %q", blink)
	}

	// Every real entry must produce a sentence that can be appended to a
	// method name, so it has to start lowercase and carry no full stop.
	for _, style := range ansistyles.All {
		text := summary(style)

		if text == "" {
			t.Errorf("summary(%q) is empty", style.Name)
			continue
		}

		if strings.HasSuffix(text, ".") {
			t.Errorf("summary(%q) ends with a full stop: %q", style.Name, text)
		}

		if upper := strings.ToUpper(text[:1]); text[:1] == upper && upper != strings.ToLower(text[:1]) {
			t.Errorf("summary(%q) starts uppercase: %q", style.Name, text)
		}
	}
}

func TestWriteDocRecordsAliases(t *testing.T) {
	var buf bytes.Buffer

	writeDoc(&buf, "Red", "paints in red", "red")

	if got, want := buf.String(), "// Red paints in red.\n"; got != want {
		t.Errorf("writeDoc for a plain style: got %q, want %q", got, want)
	}

	buf.Reset()
	writeDoc(&buf, "Gray", "paints in gray", "gray")

	if got := buf.String(); !strings.Contains(got, "It is an alias of BlackBright.") {
		t.Errorf("writeDoc for an alias: got %q, want a note naming BlackBright", got)
	}

	// underlineBlackBright is listed in aliases as the target of two other
	// names, with an empty value. It must not claim to alias anything.
	buf.Reset()
	writeDoc(&buf, "UnderlineBlackBright", "underlines in bright black", "underlineBlackBright")

	if got := buf.String(); strings.Contains(got, "alias") {
		t.Errorf("writeDoc for an alias target: got %q, want no alias note", got)
	}
}

func TestGenerateEmitsEverySpellingOfEveryStyle(t *testing.T) {
	source, err := generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// format.Source only reports an error for input it cannot parse, so
	// parsing again is the check that the emitted code is valid Go.
	if _, err := parser.ParseFile(token.NewFileSet(), "styles_gen.go", source, parser.ParseComments); err != nil {
		t.Fatalf("generated source does not parse: %v", err)
	}

	text := string(source)

	for _, style := range ansistyles.All {
		name := exported(style.Name)

		// gofmt pads the var block so that the equals signs line up, so
		// the run of spaces cannot be predicted.
		binding := regexp.MustCompile(fmt.Sprintf(`\n\tstyle%s +%s\n`,
			regexp.QuoteMeta(name), regexp.QuoteMeta(fmt.Sprintf("= mustStyle(%q)", style.Name))))
		if !binding.MatchString(text) {
			t.Errorf("generated source is missing a binding for style %q", style.Name)
		}

		for _, want := range []string{
			fmt.Sprintf("func (s Style) %s() Style {\n", name),
			fmt.Sprintf("func (i *Instance) %s() Style {\n", name),
			fmt.Sprintf("func %s() Style {\n", name),
			fmt.Sprintf("// %s %s.\n", name, summary(style)),
		} {
			if !strings.Contains(text, want) {
				t.Errorf("generated source is missing %q", want)
			}
		}
	}

	for _, model := range models {
		for _, want := range []string{
			fmt.Sprintf("func (i *Instance) %s(%s) Style {\n", model.name, model.params),
			fmt.Sprintf("func %s(%s) Style {\n", model.name, model.params),
		} {
			if !strings.Contains(text, want) {
				t.Errorf("generated source is missing %q", want)
			}
		}
	}

	for _, want := range []string{
		"func (i *Instance) Visible() Style {\n",
		"func Visible() Style {\n",
		"func (i *Instance) By(name string) (Style, error) {\n",
		"func By(name string) (Style, error) {\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("generated source is missing %q", want)
		}
	}
}

// TestGeneratedStylesAreUpToDate guards against styles_gen.go being edited by
// hand or left stale after a change to the style table or the generator. It
// drives main so that flag handling and file writing are covered too.
func TestGeneratedStylesAreUpToDate(t *testing.T) {
	committed, err := os.ReadFile(filepath.Join("..", "styles_gen.go"))
	if err != nil {
		t.Fatalf("read styles_gen.go: %v", err)
	}

	output := filepath.Join(t.TempDir(), "styles_gen.go")

	savedArgs, savedFlags := os.Args, flag.CommandLine

	t.Cleanup(func() {
		os.Args, flag.CommandLine = savedArgs, savedFlags
	})

	// The test binary owns the default flag set, so main needs a fresh one.
	flag.CommandLine = flag.NewFlagSet("gen", flag.ContinueOnError)
	os.Args = []string{"gen", "-o", output}

	main()

	fresh, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read regenerated file: %v", err)
	}

	if !bytes.Equal(fresh, committed) {
		t.Error("styles_gen.go is stale; run `go generate ./...`")
	}
}
