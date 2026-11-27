package test

import (
	"slices"
	"strings"
	"testing"
)

// Off the terminal, in a window, ? shows ed's keys as rows, which a
// window's view lays out, and esc closes them, back to the text.
func TestWindowedHelp(t *testing.T) {
	_, press := ed(t, "alpha\n")
	s := keys(press, "?")
	saves := func(l string) bool { return strings.HasSuffix(l, "  save") }
	if s.kind != "help" || !slices.ContainsFunc(s.lines, saves) {
		t.Fatalf("? drew %q %q, not ed's keys", s.kind, s.lines)
	}
	if s = keys(press, "ESC"); s.kind != "text" || s.lines[0] != "alpha" {
		t.Errorf("esc left %q %q, not the text", s.kind, s.lines)
	}
}
