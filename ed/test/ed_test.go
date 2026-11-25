// Package test drives ed as a host does, through its stream (ed -s): keys,
// clicks and drags go in as JSON lines and what it shows comes back, so a
// check reads the lines and the mode a person would see.
package test

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sys/outline"
)

// built is ed made from this tree, as ./build makes it: tangled, then
// compiled, into a folder of the test's own.
func built(t *testing.T) string {
	dir := t.TempDir()
	tangle := filepath.Join(dir, "tangle")
	run := func(in string, args ...string) {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = in
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err,
				out)
		}
	}
	run("../../tangle", "go", "build", "-o", tangle, "main.go")
	run("..", tangle)
	run("..", "go", "build", "-o", filepath.Join(dir, "ed"), "main.go")
	return filepath.Join(dir, "ed")
}

// shown is what ed drew last: its lines, its mode, and its word.
type shown struct {
	lines      []string
	mode, said string
}

// ed is ed -s on a file holding text, env the last of its environment,
// and a way to send it a key: it answers with what ed shows next. A key
// on a line names it from 1.
func ed(t *testing.T, text string, env ...string) (string,
	func(outline.Key) shown) {
	file := filepath.Join(t.TempDir(), "note")
	os.WriteFile(file, []byte(text), 0o644)
	c := exec.Command(built(t), "-s", file)
	c.Env = append(os.Environ(), "HOME="+t.TempDir(), "TMUX_PANE=")
	c.Env = append(c.Env, env...)
	in, _ := c.StdinPipe()
	out, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { in.Close(); c.Wait() })
	dec := json.NewDecoder(bufio.NewReader(out))
	enc := json.NewEncoder(in)
	next := func() shown {
		var m struct {
			Draw struct {
				Tree   outline.Node
				Status string
				State  []string
			}
		}
		if err := dec.Decode(&m); err != nil {
			t.Fatal(err)
		}
		var s shown
		for _, k := range m.Draw.Tree.Kids {
			s.lines = append(s.lines, k.Text)
		}
		s.mode, s.said = m.Draw.State[0], m.Draw.Status
		return s
	}
	next()
	return file, func(k outline.Key) shown {
		enc.Encode(map[string]outline.Key{"key": k})
		return next()
	}
}

// keys presses each of ks in turn, a letter or a name as tty spells it,
// and answers with what ed shows after the last.
func keys(press func(outline.Key) shown, ks ...string) shown {
	var s shown
	for _, k := range ks {
		s = press(outline.Key{Name: k})
	}
	return s
}

// A drag selects from where the press put the cursor to where it goes,
// stopping at a line's last rune when it goes past the end, as in any
// editor, and goes into insert, so typing replaces the selection and the
// line break stays.
func TestDrag(t *testing.T) {
	file, press := ed(t, "hello world\nsecond line\n")
	line := file + ":1"
	press(outline.Key{Node: line, At: 7})
	s := press(outline.Key{Name: "DRAG", Node: line, At: 30})
	if s.mode != "insert" {
		t.Errorf("a drag left ed in %s, not insert", s.mode)
	}
	s = keys(press, "X")
	if got := strings.Join(s.lines, "|"); got != "hello X|second line" {
		t.Errorf("typed over a drag past the end: %q", got)
	}
}

// The first undo after a drag undoes the change before it: a drag takes
// no snapshot of its own, the first edit after it does. So does an i.
func TestUndoAfterDrag(t *testing.T) {
	file, press := ed(t, "abc\n")
	keys(press, "d")
	press(outline.Key{Node: file + ":1", At: 1})
	press(outline.Key{Name: "DRAG", Node: file + ":1", At: 2})
	if s := keys(press, "C-z"); s.lines[0] != "abc" {
		t.Errorf("ctrl-z after a drag: %q, want abc", s.lines[0])
	}
	keys(press, "ESC", "i", "ESC")
	if s := keys(press, "u"); s.said != "nothing to undo" {
		t.Errorf("an i that typed nothing left an undo: %q", s.said)
	}
}

// Inside insert, ctrl-z and ctrl-y step through history one change at a
// time: what is typed after a redo is a change of its own.
func TestUndoInInsert(t *testing.T) {
	_, press := ed(t, "alpha\n")
	for _, c := range []struct{ key, want string }{
		{"i", "alpha"}, {"X", "Xalpha"}, {"C-z", "alpha"},
		{"C-y", "Xalpha"}, {"Y", "XYalpha"}, {"C-z", "Xalpha"},
	} {
		if s := keys(press, c.key); s.lines[0] != c.want {
			t.Errorf("%s: %q, want %q", c.key, s.lines[0], c.want)
		}
	}
}

// . types no insert a drag was part of: the mouse's part of it is no key.
func TestRepeatSkipsDrag(t *testing.T) {
	file, press := ed(t, "one\ntwo\nthree\n")
	keys(press, "i", "a", "b")
	press(outline.Key{Node: file + ":2", At: 1})
	press(outline.Key{Name: "DRAG", Node: file + ":2", At: 3})
	keys(press, "Z", "ESC")
	press(outline.Key{Node: file + ":3", At: 1})
	if s := keys(press, "."); s.lines[2] != "three" {
		t.Errorf(". after an insert with a drag in it: %q", s.lines[2])
	}
}

// A middle click pastes where it falls, past a line's end too, in normal
// as in insert.
func TestPasteWhereItFalls(t *testing.T) {
	file, press := ed(t, "hello world\n")
	s := press(outline.Key{Name: "PASTE", Node: file + ":1", At: 30,
		Text: "ZZ"})
	if s.lines[0] != "hello worldZZ" {
		t.Errorf("pasted past the end: %q", s.lines[0])
	}
}

// In a frame's slot q ends nothing, for it is fm's; and ed names its file
// in the window's unsaved file while it holds it unsaved, for fm's q to
// ask about: from an edit until a save, and until ed ends as its view
// goes. One that an ed killed unsaved left goes as ed starts.
func TestFrame(t *testing.T) {
	home := t.TempDir()
	mark := filepath.Join(home, ".config", "sys", "unsaved", "none")
	os.MkdirAll(filepath.Dir(mark), 0o755)
	os.WriteFile(mark, []byte("/left/by/a/killed/ed\n"), 0o644)
	named := func() string {
		b, _ := os.ReadFile(mark)
		return strings.TrimSpace(string(b))
	}
	t.Run("slot", func(t *testing.T) {
		file, press := ed(t, "abc\n", "HOME="+home, "STUDIO=ed")
		if s := keys(press, "q"); s.said != "q in fm ends the frame" {
			t.Errorf("q in a slot says %q", s.said)
		}
		if got := named(); got != "" {
			t.Errorf("ed started, and the file names %q", got)
		}
		keys(press, "i", "x", "ESC")
		if got := named(); got != file {
			t.Errorf("an edit: the file names %q, not %s",
				got, file)
		}
		keys(press, "s")
		if got := named(); got != "" {
			t.Errorf("saved, and the file names %q", got)
		}
		keys(press, "i", "y", "ESC") // left unsaved as the view goes
	})
	if got := named(); got != "" {
		t.Errorf("ed ended, and the file names %q", got)
	}
}
