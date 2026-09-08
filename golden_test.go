package chalk_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/chalk/chalk-go"
	"github.com/chalk/chalk-go/ansistyles"
)

// The golden corpus in testdata/golden.json is produced by executing the
// original JavaScript chalk. Every expectation here therefore comes from the
// implementation being ported, not from this one.
//
// The generator is migration tooling, not part of this Go module, so it lives
// outside the port. See tools/README.md for the regeneration procedure.

type goldenFile struct {
	Names struct {
		ModifierNames        []string `json:"modifierNames"`
		ForegroundColorNames []string `json:"foregroundColorNames"`
		BackgroundColorNames []string `json:"backgroundColorNames"`
		UnderlineColorNames  []string `json:"underlineColorNames"`
		ColorNames           []string `json:"colorNames"`
	} `json:"names"`

	StyleTable map[string]struct {
		Open  string `json:"open"`
		Close string `json:"close"`
	} `json:"styleTable"`

	GroupCloses struct {
		Color          string `json:"color"`
		BgColor        string `json:"bgColor"`
		UnderlineColor string `json:"underlineColor"`
	} `json:"groupCloses"`

	Convert struct {
		Ansi256ToAnsi []int            `json:"ansi256ToAnsi"`
		RGBToAnsi256  [][4]int         `json:"rgbToAnsi256"`
		RGBToAnsi     [][4]int         `json:"rgbToAnsi"`
		HexToRGB      []hexToRGBRow    `json:"hexToRgb"`
		HexToAnsi256  []hexToNumberRow `json:"hexToAnsi256"`
		HexToAnsi     []hexToNumberRow `json:"hexToAnsi"`
	} `json:"convert"`

	Cases []goldenCase `json:"cases"`
}

type hexToRGBRow struct {
	Hex string
	RGB [3]int
}

func (r *hexToRGBRow) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("hexToRgb row: want 2 elements, got %d", len(raw))
	}

	if err := json.Unmarshal(raw[0], &r.Hex); err != nil {
		return err
	}

	return json.Unmarshal(raw[1], &r.RGB)
}

type hexToNumberRow struct {
	Hex   string
	Value int
}

func (r *hexToNumberRow) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("hex row: want 2 elements, got %d", len(raw))
	}

	if err := json.Unmarshal(raw[0], &r.Hex); err != nil {
		return err
	}

	return json.Unmarshal(raw[1], &r.Value)
}

type goldenCase struct {
	ID       string `json:"id"`
	Level    int    `json:"level"`
	Ops      []op   `json:"ops"`
	Args     []any  `json:"args"`
	Expected string `json:"expected"`
}

type op struct {
	Kind string `json:"k"`
	Name string `json:"name"`
	Args []any  `json:"args"`
}

func loadGolden(t *testing.T) *goldenFile {
	t.Helper()

	data, err := os.ReadFile("testdata/golden.json")
	if err != nil {
		t.Fatalf("read golden corpus: %v", err)
	}

	var golden goldenFile
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("parse golden corpus: %v", err)
	}

	return &golden
}

// normalizeArgument converts a decoded JSON value into the Go value whose
// default formatting matches JavaScript's string conversion.
//
// JSON has a single number type, so integers arrive as float64 and would
// otherwise format via %g.
func normalizeArgument(v any) any {
	f, ok := v.(float64)
	if !ok {
		return v
	}

	if f == float64(int64(f)) {
		return int64(f)
	}

	return f
}

func toUint8(t *testing.T, v any) uint8 {
	t.Helper()

	f, ok := v.(float64)
	if !ok {
		t.Fatalf("want number, got %T", v)
	}

	return uint8(f)
}

func toString(t *testing.T, v any) string {
	t.Helper()

	s, ok := v.(string)
	if !ok {
		t.Fatalf("want string, got %T", v)
	}

	return s
}

// applyOps rebuilds a chain from the corpus description.
func applyOps(t *testing.T, style chalk.Style, ops []op) chalk.Style {
	t.Helper()

	for _, o := range ops {
		switch o.Kind {
		case "style":
			next, err := style.By(o.Name)
			if err != nil {
				t.Fatalf("By(%q): %v", o.Name, err)
			}

			style = next
		case "visible":
			style = style.Visible()
		case "rgb":
			style = style.RGB(toUint8(t, o.Args[0]), toUint8(t, o.Args[1]), toUint8(t, o.Args[2]))
		case "bgRgb":
			style = style.BgRGB(toUint8(t, o.Args[0]), toUint8(t, o.Args[1]), toUint8(t, o.Args[2]))
		case "underlineRgb":
			style = style.UnderlineRGB(toUint8(t, o.Args[0]), toUint8(t, o.Args[1]), toUint8(t, o.Args[2]))
		case "hex":
			style = style.Hex(toString(t, o.Args[0]))
		case "bgHex":
			style = style.BgHex(toString(t, o.Args[0]))
		case "underlineHex":
			style = style.UnderlineHex(toString(t, o.Args[0]))
		case "ansi256":
			style = style.Ansi256(toUint8(t, o.Args[0]))
		case "bgAnsi256":
			style = style.BgAnsi256(toUint8(t, o.Args[0]))
		case "underlineAnsi256":
			style = style.UnderlineAnsi256(toUint8(t, o.Args[0]))
		default:
			t.Fatalf("unknown op %q", o.Kind)
		}
	}

	return style
}

func TestGoldenCases(t *testing.T) {
	golden := loadGolden(t)

	if len(golden.Cases) == 0 {
		t.Fatal("golden corpus is empty")
	}

	// One instance per level, reused across cases.
	instances := map[int]*chalk.Instance{}
	for level := range 4 {
		instances[level] = chalk.MustNew(chalk.WithLevel(level))
	}

	failures := 0

	for _, c := range golden.Cases {
		instance, ok := instances[c.Level]
		if !ok {
			t.Fatalf("%s: bad level %d", c.ID, c.Level)
		}

		style := applyOps(t, instance.Root(), c.Ops)

		args := make([]any, len(c.Args))
		for i, v := range c.Args {
			args[i] = normalizeArgument(v)
		}

		got := style.Sprint(args...)

		if got != c.Expected {
			failures++
			if failures <= 25 {
				t.Errorf("%s\n  got  %q\n  want %q", c.ID, got, c.Expected)
			}
		}
	}

	if failures > 25 {
		t.Errorf("... and %d more failures", failures-25)
	}

	t.Logf("replayed %d cases from the JavaScript implementation", len(golden.Cases))
}

func TestGoldenStyleTable(t *testing.T) {
	golden := loadGolden(t)

	if len(golden.StyleTable) != len(ansistyles.All) {
		t.Fatalf("style count: got %d, want %d", len(ansistyles.All), len(golden.StyleTable))
	}

	for name, want := range golden.StyleTable {
		got, ok := ansistyles.Lookup(name)
		if !ok {
			t.Errorf("missing style %q", name)
			continue
		}

		if got.Open != want.Open {
			t.Errorf("%s open: got %q, want %q", name, got.Open, want.Open)
		}

		if got.Close != want.Close {
			t.Errorf("%s close: got %q, want %q", name, got.Close, want.Close)
		}
	}

	if ansistyles.CloseColor != golden.GroupCloses.Color {
		t.Errorf("color close: got %q, want %q", ansistyles.CloseColor, golden.GroupCloses.Color)
	}

	if ansistyles.CloseBgColor != golden.GroupCloses.BgColor {
		t.Errorf("bgColor close: got %q, want %q", ansistyles.CloseBgColor, golden.GroupCloses.BgColor)
	}

	if ansistyles.CloseUnderlineColor != golden.GroupCloses.UnderlineColor {
		t.Errorf("underlineColor close: got %q, want %q", ansistyles.CloseUnderlineColor, golden.GroupCloses.UnderlineColor)
	}
}

func TestGoldenNameLists(t *testing.T) {
	golden := loadGolden(t)

	lists := []struct {
		name string
		got  []string
		want []string
	}{
		{"ModifierNames", ansistyles.ModifierNames, golden.Names.ModifierNames},
		{"ForegroundColorNames", ansistyles.ForegroundColorNames, golden.Names.ForegroundColorNames},
		{"BackgroundColorNames", ansistyles.BackgroundColorNames, golden.Names.BackgroundColorNames},
		{"UnderlineColorNames", ansistyles.UnderlineColorNames, golden.Names.UnderlineColorNames},
		{"ColorNames", ansistyles.ColorNames, golden.Names.ColorNames},
	}

	for _, list := range lists {
		if len(list.got) != len(list.want) {
			t.Errorf("%s: got %d names, want %d", list.name, len(list.got), len(list.want))
			continue
		}

		// Order is part of the contract: callers iterate these.
		for i := range list.want {
			if list.got[i] != list.want[i] {
				t.Errorf("%s[%d]: got %q, want %q", list.name, i, list.got[i], list.want[i])
			}
		}
	}
}

func TestGoldenConversions(t *testing.T) {
	golden := loadGolden(t)

	for code, want := range golden.Convert.Ansi256ToAnsi {
		if got := ansistyles.Ansi256ToAnsi(uint8(code)); got != want {
			t.Errorf("Ansi256ToAnsi(%d): got %d, want %d", code, got, want)
		}
	}

	for _, row := range golden.Convert.RGBToAnsi256 {
		r, g, b, want := uint8(row[0]), uint8(row[1]), uint8(row[2]), uint8(row[3])
		if got := ansistyles.RGBToAnsi256(r, g, b); got != want {
			t.Errorf("RGBToAnsi256(%d,%d,%d): got %d, want %d", r, g, b, got, want)
		}
	}

	for _, row := range golden.Convert.RGBToAnsi {
		r, g, b, want := uint8(row[0]), uint8(row[1]), uint8(row[2]), row[3]
		if got := ansistyles.RGBToAnsi(r, g, b); got != want {
			t.Errorf("RGBToAnsi(%d,%d,%d): got %d, want %d", r, g, b, got, want)
		}
	}

	for _, row := range golden.Convert.HexToRGB {
		r, g, b := ansistyles.HexToRGB(row.Hex)
		if int(r) != row.RGB[0] || int(g) != row.RGB[1] || int(b) != row.RGB[2] {
			t.Errorf("HexToRGB(%q): got (%d,%d,%d), want %v", row.Hex, r, g, b, row.RGB)
		}
	}

	for _, row := range golden.Convert.HexToAnsi256 {
		if got := ansistyles.HexToAnsi256(row.Hex); int(got) != row.Value {
			t.Errorf("HexToAnsi256(%q): got %d, want %d", row.Hex, got, row.Value)
		}
	}

	for _, row := range golden.Convert.HexToAnsi {
		if got := ansistyles.HexToAnsi(row.Hex); got != row.Value {
			t.Errorf("HexToAnsi(%q): got %d, want %d", row.Hex, got, row.Value)
		}
	}
}
