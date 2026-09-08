package chalk

import "strings"

// Ported from source/utilities.js.
//
// Both helpers work on raw bytes. Every delimiter they search for is ASCII
// (ESC, CR, LF, and the ASCII-only SGR sequences), and UTF-8 guarantees that
// an ASCII byte never occurs inside a multi-byte rune. Decoding to runes would
// therefore be pure overhead, and would also change the index arithmetic.

// stringReplaceAll inserts postfix after every occurrence of substring.
//
// The occurrence itself is preserved. Scanning resumes after the matched text,
// so a postfix that happens to contain substring is never rescanned.
func stringReplaceAll(s, substring, postfix string) string {
	// An empty substring would match endlessly at the same offset. The
	// JavaScript original loops forever here; callers only ever pass a close
	// sequence, which is never empty.
	if substring == "" {
		return s
	}

	index := strings.Index(s, substring)
	if index == -1 {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s) + len(postfix))

	endIndex := 0

	for index != -1 {
		sb.WriteString(s[endIndex:index])
		sb.WriteString(substring)
		sb.WriteString(postfix)

		endIndex = index + len(substring)

		next := strings.Index(s[endIndex:], substring)
		if next == -1 {
			break
		}

		index = endIndex + next
	}

	sb.WriteString(s[endIndex:])

	return sb.String()
}

// stringEncaseCRLFWithFirstIndex closes a style before every line break and
// reopens it afterwards, so that styling does not bleed across lines.
//
// index must be the offset of the first '\n' in s. A '\r' immediately before a
// '\n' is treated as part of the break and is emitted after the prefix, keeping
// CRLF sequences intact.
func stringEncaseCRLFWithFirstIndex(s, prefix, postfix string, index int) string {
	var sb strings.Builder
	sb.Grow(len(s) + len(prefix) + len(postfix))

	endIndex := 0

	for {
		gotCR := index > 0 && s[index-1] == '\r'

		if gotCR {
			sb.WriteString(s[endIndex : index-1])
			sb.WriteString(prefix)
			sb.WriteString("\r\n")
		} else {
			sb.WriteString(s[endIndex:index])
			sb.WriteString(prefix)
			sb.WriteString("\n")
		}

		sb.WriteString(postfix)

		endIndex = index + 1

		next := strings.IndexByte(s[endIndex:], '\n')
		if next == -1 {
			break
		}

		index = endIndex + next
	}

	sb.WriteString(s[endIndex:])

	return sb.String()
}
