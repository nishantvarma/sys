package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

// lensOn is studio show, args after it, at the end of a pipe in dir, as a
// host runs it, with home its home and no tmux to reach: a way to press a
// key, and until, the first draw ok takes within a few seconds, which
// fails the test, saying why, when none does; ended is closed as the
// lens ends.
func lensOn(t *testing.T, home, dir string, args ...string) (
	press func(outline.Key), until func(string, func(drawn) bool) drawn,
	ended <-chan struct{}) {
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
	draws, done := make(chan drawn, 64), make(chan struct{})
	go func() {
		defer close(done)
		dec := json.NewDecoder(out)
		for {
			var m struct{ Draw *drawn }
			if dec.Decode(&m) != nil {
				return
			}
			if m.Draw != nil {
				draws <- *m.Draw
			}
		}
	}()
	enc := json.NewEncoder(in)
	press = func(k outline.Key) {
		enc.Encode(map[string]outline.Key{"key": k})
	}
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

// shows is whether a draw shows the lens named, first in its state.
func shows(lens string) func(d drawn) bool {
	return func(d drawn) bool {
		return len(d.State) > 0 && d.State[0] == lens
	}
}

// note is a note in dir that draws as a picture: two lanes, a box in
// each, and a message between them; and its rows as the outline shows
// them.
const note = "t\nup\n  a\n    → b  go\nlow\n  b\n"
const noted = "t|up|  a|    → b  go|low|  b"

// TestLens: given no lens, studio show shows every way: the note cur names
// as its outline, a row an item, the lens shown in its state, the cursor
// on the item cur's line is in. While ed holds the note unsaved it shows
// the text ed holds, and the file again once ed lets it go. Tab turns to
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
	if i := slices.IndexFunc(d.Tree.Kids,
		func(k *outline.Node) bool { return k.Cur }); i != 3 {
		t.Errorf("the cursor is on row %d, not on cur's line, 4", i)
	}
	put("text", note+"  c\n")
	put("unsaved", file+"\n")
	d = until("no text as ed holds it", func(d drawn) bool {
		return strings.HasSuffix(d.rows(), "|  c")
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
