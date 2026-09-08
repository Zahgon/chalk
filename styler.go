package chalk

import "github.com/chalk/chalk-go/ansistyles"

// styler is one link in a chain of applied styles.
//
// Each link stores its own sequences plus the fully composed sequences for the
// whole chain, so rendering never has to walk the chain. The chain itself is
// retained because nested styles are unwound link by link.
//
// A styler is immutable once built, which is what makes [Style] safe to share.
type styler struct {
	open  string
	close string

	// openAll is every open sequence from the root down to this link.
	openAll string
	// closeAll is every close sequence from this link back up to the root.
	closeAll string

	parent *styler
}

// newStyler pushes a style onto parent's chain.
//
// Opens accumulate outermost-first and closes innermost-first, so that the
// sequences nest properly: openAll grows on the right, closeAll on the left.
func newStyler(open, close string, parent *styler) *styler {
	s := &styler{
		open:   open,
		close:  close,
		parent: parent,
	}

	if parent == nil {
		s.openAll = open
		s.closeAll = close
	} else {
		s.openAll = parent.openAll + open
		s.closeAll = close + parent.closeAll
	}

	return s
}

// lookupStyle resolves a style name from the ANSI style table.
func lookupStyle(name string) (ansistyles.Style, bool) {
	return ansistyles.Lookup(name)
}
