package chalk

import "testing"

// stringReplaceAll and stringEncaseCRLFWithFirstIndex are only ever called by
// render with a real style's escape sequences, so the degenerate inputs the
// JavaScript originals guard against are unreachable from the public API.

func TestStringReplaceAllLeavesTheStringAloneForAnEmptySubstring(t *testing.T) {
	if got := stringReplaceAll("abc", "", "-"); got != "abc" {
		t.Errorf("empty substring: got %q, want %q", got, "abc")
	}

	if got := stringReplaceAll("", "", "-"); got != "" {
		t.Errorf("empty string and substring: got %q, want an empty string", got)
	}
}

func TestStringReplaceAllKeepsTheMatchAndAppendsThePostfix(t *testing.T) {
	cases := []struct {
		text, substring, postfix, want string
	}{
		{"abc", "b", "B", "abBc"},
		{"aaa", "a", "x", "axaxax"},
		{"abc", "z", "Z", "abc"},
		{"aa", "aa", "b", "aab"},
		// A postfix containing the substring must not be rescanned.
		{"ab", "a", "a", "aab"},
	}

	for _, c := range cases {
		if got := stringReplaceAll(c.text, c.substring, c.postfix); got != c.want {
			t.Errorf("stringReplaceAll(%q, %q, %q): got %q, want %q",
				c.text, c.substring, c.postfix, got, c.want)
		}
	}
}
