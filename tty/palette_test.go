package tty

import (
	"os"
	"regexp"
	"testing"
)

// The palette's page colours are the page's own tokens in their dark
// form, as the skeleton writes them: a token changed there and not here
// shows.
func TestPalette(t *testing.T) {
	b, err := os.ReadFile("../medium/page.html")
	if err != nil {
		t.Fatal(err)
	}
	tok := regexp.MustCompile(
		`--([a-z]+): light-dark\(#[0-9a-f]{6}, (#[0-9a-f]{6})\)`)
	dark := map[string]string{}
	for _, m := range tok.FindAllStringSubmatch(string(b), -1) {
		dark[m[1]] = m[2]
	}
	for token, ours := range map[string]string{"bg": Ground,
		"panel": Panel, "fg": Fore, "dim": Soft, "line": Rule,
		"accent": Accent, "plus": Plus, "minus": Minus} {
		if dark[token] != ours {
			t.Errorf("--%s: the page's %q, ours %q", token,
				dark[token], ours)
		}
	}
}

// Paint says its colours as 24-bit SGR, and ends them with the text.
func TestPaint(t *testing.T) {
	got := Paint("x", Accent, Bar, true)
	want := "\x1b[0;1;38;2;106;162;242;48;2;28;42;61mx\x1b[0m"
	if got != want {
		t.Errorf("Paint: %q, want %q", got, want)
	}
	if w := Width(got); w != 1 {
		t.Errorf("Paint's escapes take %d columns, want none", w-1)
	}
}
