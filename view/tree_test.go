package view

import (
	"regexp"
	"strings"
	"testing"

	"sys/outline"
	"sys/tty"
)

// spoken is a note, as the lens sends one: its parts read by the note
// reader, under a tree of kind, the cursor on the part at line at.
func spoken(t *testing.T, kind string, at int) *outline.Node {
	t.Helper()
	src := "view: a program and its view\n" +
		"\n" +
		"A program says what to show, as `view/term.go` does.\n" +
		"\n" +
		"Messages\n" +
		"  plain data, a line each\n" +
		"  to the view  draw, ask\n" +
		"Who holds what\n" +
		"  program  the data and the cursor\n" +
		"  between  ideas, by name\n"
	parts, err := outline.Parts(strings.NewReader(src), "n")
	if err != nil {
		t.Fatal(err)
	}
	var walk func(ns []*outline.Node)
	walk = func(ns []*outline.Node) {
		for _, n := range ns {
			if n.ID == outline.At("n", at) && !n.Cur {
				n.Cur = true
				return
			}
			walk(n.Kids)
		}
	}
	walk(parts)
	return &outline.Node{ID: "n", Kind: kind, Kids: parts}
}

// shown is the lines a view drew last, each its text, its blanks at the
// right let go.
func shown(v *Term) []string {
	var out []string
	for _, l := range v.Lines() {
		var b strings.Builder
		for _, r := range l {
			b.WriteString(r.Text)
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return out
}

// TestOutline: an outline is rows that fold. Every part is unfolded at
// first, so the rows run down to the leaves, each part saying how many
// rows it holds, and the cursor's the first of what it says too; a guide
// stands under a part's caret beside the rows it holds; a row's name
// stands in a column with its fellows'; and a blank line stands between
// parts of two kinds, and around a part that holds rows.
// ← folds an unfolded part and → unfolds it again, both the view's; →
// on an unfolded one steps in, which the program hears; ← on a part
// holding none goes to the part it is under and folds it; and as thumbs
// → goes a card on, as it does in a folder.
func TestOutline(t *testing.T) {
	v := New(nil, "")
	v.In(0, 0, 60, 18)
	v.Draw(spoken(t, "outline", 10), "hi", nil)
	got := strings.Join(shown(v)[5:15], "\n")
	want := strings.Join([]string{
		"  1    view: a program and its view",
		"",
		"  2    A program says what to show, as view/term.go does.",
		"",
		"  3  ▾ Messages                                      1 row",
		"  4  │   to the view  draw, ask",
		"",
		"  5  ▾ Who holds what                                2 rows",
		"  6  │   program  the data and the cursor",
		"▌ 7  │   between  ideas, by name",
	}, "\n")
	if got != want {
		t.Errorf("the outline:\n%s\nwant:\n%s", got, want)
	}
	row := func(of string) string { // the body's line with of in it
		for _, l := range shown(v)[1:] {
			if strings.Contains(l, of) {
				return l
			}
		}
		return ""
	}
	if !strings.Contains(shown(v)[0], ": between") {
		t.Errorf("the head names no cursor's row: %q", shown(v)[0])
	}
	for range 4 {
		v.Handle(tty.Key{Name: tty.KUp})
	}
	if l := row("Messages"); !strings.Contains(l, "plain data, a line") {
		t.Errorf("the cursor's row says no lead: %q", l)
	}
	if ev, theirs := v.Handle(tty.Key{Name: tty.KLeft}); theirs {
		t.Errorf("← on an unfolded part is the view's, not %v", ev)
	}
	if !strings.Contains(row("Messages"), "▸") ||
		row("to the view") != "" {
		t.Errorf("← folded nothing:\n%s", strings.Join(shown(v), "\n"))
	}
	if ev, theirs := v.Handle(tty.Key{Name: tty.KRight}); theirs {
		t.Errorf("→ on a folded part is the view's, not %v", ev)
	}
	if !strings.Contains(row("Messages"), "▾") ||
		row("to the view  draw, ask") == "" {
		t.Errorf("→ opened nothing:\n%s", strings.Join(shown(v), "\n"))
	}
	ev, _ := v.Handle(tty.Key{Name: tty.KRight})
	if ev.Node != outline.At("n", 7) {
		t.Errorf("→ on an open part went to %q, not its first row",
			ev.Node)
	}
	ev, _ = v.Handle(tty.Key{Name: tty.KLeft})
	if ev.Node != outline.At("n", 5) || strings.Contains(
		strings.Join(shown(v), "\n"), "to the view") {
		t.Errorf("← on a row went to %q, and its part stays open",
			ev.Node)
	}
	v.Handle(tty.Key{Rune: 't'})
	if ev, _ := v.Handle(tty.Key{Name: tty.KRight}); ev.Node == "" {
		t.Error("→ as thumbs folded, and went no card on")
	}
}

// TestDoc: a doc is set for the eye: prose filled to the width, and no
// wider than the measure however wide the view; a row's text beside its
// name, bold, filled from its column; an address underlined; the part at
// the cursor marked, its lines on the bar behind an edge; a word a click
// falls on is said whole, an address too; G scrolls to the end; and a
// scroll leaves the doc's lines as they were, one count of those below.
func TestDoc(t *testing.T) {
	v := New(nil, "")
	v.In(0, 0, 36, 24)
	v.Draw(spoken(t, "doc", 10), "hi", nil)
	lines := shown(v)
	body := strings.Join(lines[5:22], "\n")
	for _, want := range []string{
		"  A program says what to show, as\n  view/term.go does.",
		"    program  the data and the\n             cursor",
		"▌   between  ideas, by name",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the doc holds no %q:\n%s", want, body)
		}
	}
	for y, l := range lines {
		if x := strings.Index(l, "view/term.go"); x >= 0 {
			at := len([]rune(l[:x])) + 2
			if w := v.word(at, y); w != "view/term.go" {
				t.Errorf("a click on the address says %q", w)
			}
		}
	}
	if got := v.word(3, 0); got != "" {
		t.Errorf("a click on the head says %q, not its place", got)
	}
	looks := map[string]Run{}
	for _, l := range v.Lines() {
		for _, r := range l {
			looks[r.Text] = r
		}
	}
	if !looks["view/term.go"].Under || !looks["program"].Bold {
		t.Errorf("an address not underlined, a name not bold: %+v %+v",
			looks["view/term.go"], looks["program"])
	}
	v.In(0, 0, 120, 24)
	for _, l := range shown(v)[5:22] {
		if n := len([]rune(l)); n > margin+measure {
			t.Errorf("a line %d wide, past the measure: %q", n, l)
		}
	}
	v.In(0, 0, 36, 12)
	v.Handle(tty.Key{Rune: 'g'})
	top := strings.Join(shown(v), "\n")
	if !strings.Contains(top, "view: a program") {
		t.Errorf("g scrolled not to the top:\n%s", top)
	}
	count := regexp.MustCompile(`^ \+[0-9]+$`) // how many lines below
	for i := 0; i < 4; i++ {
		v.Handle(tty.Key{Name: tty.KDown})
		n := 0 // drawn at each step, as a window draws a frame
		for _, l := range shown(v) {
			if count.MatchString(l) {
				n++
			}
		}
		if n > 1 {
			t.Fatalf("a count among the doc's lines, %d down:\n%s",
				i+1, strings.Join(shown(v), "\n"))
		}
	}
	v.Handle(tty.Key{Rune: 'G'})
	end := strings.Join(shown(v), "\n")
	if strings.Contains(end, "view: a program") ||
		!strings.Contains(end, "ideas, by name") {
		t.Errorf("G scrolled not to the end:\n%s", end)
	}
}

// TestDocDeclared: a declaration stands on the panel, its name bold, and
// its doc hangs under it, four columns in, its summary bright and its
// later paragraphs soft.
func TestDocDeclared(t *testing.T) {
	src := "package p\n\n// F says what it does.\n//\n" +
		"// And how, in more words.\nfunc F() {}\n"
	parts, err := outline.Go(strings.NewReader(src), "p.go")
	if err != nil {
		t.Fatal(err)
	}
	v := New(nil, "")
	v.In(0, 0, 60, 20)
	v.Draw(&outline.Node{ID: "p.go", Kind: "doc", Kids: parts}, "", nil)
	looks, at := map[string]Run{}, map[string]int{}
	for y, l := range v.Lines() {
		line := shown(v)[y]
		said := strings.TrimSpace(line)
		for _, r := range l {
			if strings.HasPrefix(said, "func ") || r.Text != "F" {
				looks[r.Text] = r // the name's run, its line's
			}
		}
		at[said] = len(line) - len(strings.TrimLeft(line, " "))
	}
	if n := at["F says what it does."]; n != margin+4 {
		t.Errorf("the doc stands at column %d, not %d", n, margin+4)
	}
	name, later := looks["F"], looks["And"]
	if !name.Bold || name.Ground != tty.Panel || later.Color != tty.Soft {
		t.Errorf("the name %+v, the later paragraph %+v", name, later)
	}
}
