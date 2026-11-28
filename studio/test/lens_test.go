package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sys/outline"
)

// drawn is one draw of the lens's, as its view hears it.
type drawn struct {
	Tree   outline.Node
	Status string
	State  []string
}

// rows is a drawing's rows, each its text.
func (d drawn) rows() string {
	var out []string
	for _, k := range d.Tree.Kids {
		out = append(out, k.Text)
	}
	return strings.Join(out, "|")
}

// find is the first part of a drawing, its parts and theirs, that ok
// takes; nil for none.
func (d drawn) find(ok func(*outline.Node) bool) *outline.Node {
	var walk func(ns []*outline.Node) *outline.Node
	walk = func(ns []*outline.Node) *outline.Node {
		for _, n := range ns {
			if ok(n) {
				return n
			}
			if f := walk(n.Kids); f != nil {
				return f
			}
		}
		return nil
	}
	return walk(d.Tree.Kids)
}

// lensOn is studio show, args after it, at the end of a pipe in dir, as a
// host runs it, with home its home and no tmux to reach: a way to press a
// key, and until, the first draw ok takes within a few seconds, which
// fails the test, saying why, when none does; ended is closed as the
// lens ends.
func lensOn(t *testing.T, home, dir string, args ...string) (
	press func(outline.Key), until func(string, func(drawn) bool) drawn,
	ended <-chan struct{}) {
	press, draws, done := lensDraws(t, home, dir, args...)
	var last drawn
	until = func(why string, ok func(d drawn) bool) drawn {
		t.Helper()
		end := time.After(5 * time.Second)
		for {
			select {
			case d := <-draws:
				if last = d; ok(d) {
					return d
				}
			case <-done:
				t.Fatalf("%s: the lens ended", why)
			case <-end:
				t.Fatalf("%s: last drew %q %v %q", why,
					last.Tree.Kind, last.State,
					last.rows())
			}
		}
	}
	return press, until, done
}

// lensDraws is studio show, as lensOn runs it: a way to press a key, and
// each draw as it comes; done is closed as the lens ends.
func lensDraws(t *testing.T, home, dir string, args ...string) (
	press func(outline.Key), draws <-chan drawn, done <-chan struct{}) {
	c := exec.Command(built(t), append([]string{"show"}, args...)...)
	c.Env = append(os.Environ(), "HOME="+home, "TMUX=", "TMUX_PANE=",
		"STUDIO=")
	c.Dir = dir
	in, _ := c.StdinPipe()
	out, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { in.Close(); c.Process.Kill(); c.Wait() })
	ds, ended := make(chan drawn, 64), make(chan struct{})
	go func() {
		defer close(ended)
		dec := json.NewDecoder(out)
		for {
			var m struct{ Draw *drawn }
			if dec.Decode(&m) != nil {
				return
			}
			if m.Draw != nil {
				ds <- *m.Draw
			}
		}
	}()
	enc := json.NewEncoder(in)
	press = func(k outline.Key) {
		enc.Encode(map[string]outline.Key{"key": k})
	}
	return press, ds, ended
}

// shows is whether a draw shows the lens named, first in its state.
func shows(lens string) func(d drawn) bool {
	return func(d drawn) bool {
		return len(d.State) > 0 && d.State[0] == lens
	}
}

// note is a note in dir that draws as a picture: two lanes, a box in
// each, and a message between them; and its parts as the outline shows
// them at its top, its title, run on as a paragraph does, and the
// sections under it (`outline/parts.go:/func Parts/`).
const note = "t\nup\n  a\n    → b  go\nlow\n  b\n"
const noted = "t up|a|low"

// TestLens: given no lens, studio show shows every way: the note cur names
// as its outline, a tree of its parts, the lens shown in its state, the
// cursor on the part cur's line is in. While ed holds the note unsaved it
// shows the text ed holds, and the file again once ed lets it go. Tab turns to
// the render, the note's items as they nest for its picture, and again to
// the diff; git reads the file as saved, and the diff says so while ed
// holds the note. A key on a row, once the pane has the eye, says the row
// to cur; and q, outside a frame, ends it.
func TestLens(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	file := filepath.Join(dir, "note")
	os.WriteFile(file, []byte(note), 0o644)
	state := filepath.Join(home, ".config", "sys")
	put := func(kind, s string) {
		p := filepath.Join(state, kind, "none")
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p+".new", []byte(s), 0o644)
		os.Rename(p+".new", p)
	}
	read := func(kind string) string {
		b, _ := os.ReadFile(filepath.Join(state, kind, "none"))
		return strings.TrimSpace(string(b))
	}
	put("cur", file+":4\n")
	press, until, ended := lensOn(t, home, dir, "-s")

	d := until("no outline", shows("outline"))
	if d.Tree.Kind != "outline" || d.rows() != noted {
		t.Fatalf("the outline: %q %q", d.Tree.Kind, d.rows())
	}
	marked := func(k *outline.Node) bool { return k.Cur }
	if c := d.find(marked); c == nil || c.ID != outline.At(file, 4) {
		t.Errorf("the cursor is on %v, not on cur's line, 4", c)
	}
	put("text", note+"  c\n")
	put("unsaved", file+"\n")
	d = until("no text as ed holds it", func(d drawn) bool {
		return d.find(func(k *outline.Node) bool {
			return k.Text == "b c"
		}) != nil
	})
	if !strings.Contains(d.Status, "unsaved") {
		t.Errorf("ed's text, and the word is %q", d.Status)
	}
	os.Remove(filepath.Join(state, "unsaved", "none"))
	until("ed let go, and the file is not shown", func(d drawn) bool {
		return d.rows() == noted
	})
	press(outline.Key{Name: "TAB"})
	d = until("Tab showed no render", shows("render"))
	if d.Tree.Kind != "picture" || d.rows() != "t|up|low" {
		t.Errorf("the render: %q %q", d.Tree.Kind, d.rows())
	}
	press(outline.Key{Name: "TAB"})
	d = until("Tab again showed no diff", shows("diff"))
	if len(d.Tree.Kids) != 1 || d.Tree.Kids[0].Kind != "waits" {
		t.Errorf("a diff outside git says no why: %q", d.rows())
	}
	put("unsaved", file+"\n")
	saved := func(d drawn) bool {
		return strings.HasSuffix(d.Status, "of the file as saved")
	}
	until("the diff, ed holding the note, says nothing of it", saved)
	os.Remove(filepath.Join(state, "unsaved", "none"))
	press(outline.Key{Name: "TAB"})
	until("Tab from the last lens is not the first", shows("outline"))
	press(outline.Key{Name: "FOCUS"})
	press(outline.Key{Node: outline.At(file, 5)})
	until("a key on a row", func(drawn) bool {
		return read("cur") == outline.At(file, 5)
	})
	press(outline.Key{Name: "q"})
	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("q outside a frame did not end the lens")
	}
}

// TestLensOne: given a lens, studio show shows that one alone, as a tab
// of a frame's lens view does: Tab is the frame's, so outside a frame it
// says why and shows the same lens.
func TestLensOne(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	file := filepath.Join(dir, "note")
	os.WriteFile(file, []byte(note), 0o644)
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	os.WriteFile(cur, []byte(file+"\n"), 0o644)
	press, until, _ := lensOn(t, home, dir, "render", "-s")
	d := until("no render", shows("render"))
	if d.Tree.Kind != "picture" {
		t.Fatalf("the render alone: %q", d.Tree.Kind)
	}
	press(outline.Key{Name: "TAB"})
	d = until("Tab outside a frame said nothing", func(d drawn) bool {
		return d.Status != "" && !strings.HasPrefix(d.Status, "what")
	})
	if !shows("render")(d) {
		t.Errorf("Tab outside a frame turned the lens: %v", d.State)
	}
}

// TestLensKinds: the lens reads each thing by the reader for its kind
// (`studio/4-lens/26-readers.go`): a Go file by its declarations, which
// render as a doc; a note with no message, a doc too; a script by its
// indent, a line a part, which has no render but its source, and says so.
func TestLensKinds(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	files := map[string]string{
		"f.go":   "package f\n\n// F is one.\nfunc F() {}\n",
		"note":   "a note\n\nprose, and no message\n",
		"script": "#!/bin/sh\necho hi\n",
	}
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	for name, text := range files {
		os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644)
	}
	kinds := func(d drawn) string {
		var out []string
		for _, k := range d.Tree.Kids {
			out = append(out, k.Kind)
		}
		return strings.Join(out, " ")
	}
	for _, c := range []struct{ file, parts, render string }{
		{"f.go", "title func", "doc"},
		{"note", "title para", "doc"},
		{"script", "line line", "waits"},
	} {
		at := filepath.Join(dir, c.file) + "\n"
		os.WriteFile(cur, []byte(at), 0o644)
		_, until, _ := lensOn(t, home, dir, "outline", "-s")
		d := until(c.file+": no outline", shows("outline"))
		if d.Tree.Kind != "outline" || kinds(d) != c.parts {
			t.Errorf("%s: the outline is %q, its parts %q, not %q",
				c.file, d.Tree.Kind, kinds(d), c.parts)
		}
		_, until, _ = lensOn(t, home, dir, "render", "-s")
		d = until(c.file+": no render", shows("render"))
		got := d.Tree.Kind // a render that says why is no kind, a why
		if got == "" {
			got = kinds(d)
		}
		if got != c.render {
			t.Errorf("%s: the render is %q, not %q", c.file, got,
				c.render)
		}
	}
}

// TestLensPython: Python's reader, page/py -t, gives each part the line
// it starts on, as its id, a docstring's every paragraph too, so that no
// two parts share one and the cursor lights the one it is on; and a def's
// name is a run of its text, not a part under it.
func TestLensPython(t *testing.T) {
	src := `"""Title.

Two.

Three,
run on.

Four.
"""


def f():
    pass
`
	c := exec.Command("../../medium/page/py", "-t", "m.py")
	c.Stdin = strings.NewReader(src)
	out, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	var parts []*outline.Node
	if err := json.Unmarshal(out, &parts); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, p := range parts {
		ids = append(ids, p.ID)
	}
	want := "m.py:1 m.py:3 m.py:5 m.py:8 m.py:12"
	if got := strings.Join(ids, " "); got != want {
		t.Errorf("the parts start at %s, not %s", got, want)
	}
	f := parts[len(parts)-1]
	if len(f.Runs) != 3 || f.Runs[1].Kind != "name" ||
		f.Runs[1].Text != "f" || len(f.Kids) != 0 {
		t.Errorf("def f(): runs %v, parts %v", f.Runs, f.Kids)
	}
}

// TestLensGo: Go's reader puts a method under its type, out of the order
// the file has them, and the cursor still goes to the part ed's line is
// in, the method; and a file half typed shows as far as it parses, the
// word saying where it stopped.
func TestLensGo(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	m := filepath.Join(dir, "m.go")
	os.WriteFile(m, []byte("package m\n\ntype T struct{}\n\n"+
		"func F() {}\n\nfunc (T) M() {\n\treturn\n}\n"), 0o644)
	b := filepath.Join(dir, "b.go")
	os.WriteFile(b, []byte("package b\n\nfunc F() {\n"), 0o644)
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	os.WriteFile(cur, []byte(m+":8\n"), 0o644)
	_, until, _ := lensOn(t, home, dir, "outline", "-s")
	d := until("no outline of m.go", shows("outline"))
	marked := func(k *outline.Node) bool { return k.Cur }
	if c := d.find(marked); c == nil || c.ID != outline.At(m, 7) {
		t.Errorf("the cursor is on %v, not on M, line 7", c)
	}
	os.WriteFile(cur, []byte(b+"\n"), 0o644)
	_, until, _ = lensOn(t, home, dir, "outline", "-s")
	d = until("no outline of b.go", shows("outline"))
	if len(d.Tree.Kids) != 2 || !strings.Contains(d.Status, "b.go:3:") {
		t.Errorf("a file half typed: %d parts, the word %q",
			len(d.Tree.Kids), d.Status)
	}
}

// TestLensDrawsOnce: a key on a row draws the lens once, the cursor there,
// and nothing after it: its own write to cur, which wakes it, changes
// nothing it shows, so it draws nothing again.
func TestLensDrawsOnce(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	file := filepath.Join(dir, "note")
	os.WriteFile(file, []byte(note), 0o644)
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	os.WriteFile(cur, []byte(file+":1\n"), 0o644)
	press, draws, _ := lensDraws(t, home, dir, "outline", "-s")
	<-draws
	press(outline.Key{Name: "FOCUS"})
	press(outline.Key{Node: outline.At(file, 5)})
	n := 0
	for quiet := time.After(time.Second); ; {
		select {
		case <-draws:
			n++
			continue
		case <-quiet:
		}
		break
	}
	if n != 1 {
		t.Errorf("a key on a row drew %d times, not once", n)
	}
}
