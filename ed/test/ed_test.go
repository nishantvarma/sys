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
	"time"

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

// shown is what ed drew last: its lines, its mode, and its word; and the
// kind of tree it drew, a text or its keys.
type shown struct {
	lines      []string
	mode, said string
	kind       string
}

// ed is ed -s on a file holding text, env the last of its environment,
// and a way to send it a key: it answers with what ed shows next. A key
// on a line names it from 1.
func ed(t *testing.T, text string, env ...string) (string,
	func(outline.Key) shown) {
	file := filepath.Join(t.TempDir(), "note")
	os.WriteFile(file, []byte(text), 0o644)
	return file, on(t, file, env...)
}

// on is ed -s on the file at path, as it is, and a way to send it a key.
func on(t *testing.T, file string, env ...string) func(outline.Key) shown {
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
		s.kind = m.Draw.Tree.Kind
		return s
	}
	next()
	return func(k outline.Key) shown {
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
// editor, and leaves ed in the mode it was in: in normal the selection is
// the verbs', so c types over it and the line break stays, and . does the
// same to the next selection, as after one made by keys.
func TestDrag(t *testing.T) {
	file, press := ed(t, "hello world\nsecond line\n")
	line := file + ":1"
	press(outline.Key{Node: line, At: 7})
	s := press(outline.Key{Name: "DRAG", Node: line, At: 30})
	if s.mode != "normal" {
		t.Errorf("a drag left ed in %s, not normal", s.mode)
	}
	s = keys(press, "c", "X", "ESC")
	if got := strings.Join(s.lines, "|"); got != "hello X|second line" {
		t.Errorf("changed a drag past the end: %q", got)
	}
	press(outline.Key{Node: file + ":2", At: 1})
	press(outline.Key{Name: "DRAG", Node: file + ":2", At: 6})
	s = keys(press, ".")
	if got := strings.Join(s.lines, "|"); got != "hello X|X line" {
		t.Errorf(". after a change to a drag: %q", got)
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
// ask about, and puts the text it holds in the window's text file, for
// the lens to read: from an edit until a save, and until ed ends as its
// view goes. What an ed killed unsaved left goes as ed starts.
func TestFrame(t *testing.T) {
	home := t.TempDir()
	mark := filepath.Join(home, ".config", "sys", "unsaved", "none")
	text := filepath.Join(home, ".config", "sys", "text", "none")
	for _, p := range []string{mark, text} {
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte("/left/by/a/killed/ed\n"), 0o644)
	}
	named := func() string {
		b, _ := os.ReadFile(mark)
		return strings.TrimSpace(string(b))
	}
	held := func() string {
		b, _ := os.ReadFile(text)
		return string(b)
	}
	t.Run("slot", func(t *testing.T) {
		file, press := ed(t, "abc\n", "HOME="+home, "STUDIO=ed")
		if s := keys(press, "q"); s.said != "q in fm ends the frame" {
			t.Errorf("q in a slot says %q", s.said)
		}
		if got := named() + held(); got != "" {
			t.Errorf("ed started, and the files hold %q", got)
		}
		keys(press, "i", "x", "ESC")
		if got := named(); got != file {
			t.Errorf("an edit: the file names %q, not %s",
				got, file)
		}
		if got := held(); got != "xabc\n" {
			t.Errorf("an edit: the text file holds %q", got)
		}
		keys(press, "i", "y", "ESC")
		if got := held(); got != "yxabc\n" {
			t.Errorf("another: the text file holds %q", got)
		}
		keys(press, "s")
		if got := named() + held(); got != "" {
			t.Errorf("saved, and the files hold %q", got)
		}
		keys(press, "i", "y", "ESC") // left unsaved as the view goes
	})
	if got := named() + held(); got != "" {
		t.Errorf("ed ended, and the files hold %q", got)
	}
}

// A save puts the file whole, written beside it and renamed over it, and
// leaves nothing else beside it: a link stays a link, the file it names
// changing, and the file keeps its mode; a file with a second name is
// written in place, so the other name reads the new text too.
func TestSave(t *testing.T) {
	dir := t.TempDir()
	note := filepath.Join(dir, "note")
	os.WriteFile(note, []byte("a\n"), 0o750)
	link := filepath.Join(dir, "link")
	os.Symlink("note", link)
	was, _ := os.Stat(note)
	keys(on(t, link), "i", "x", "ESC", "s")
	if now, _ := os.Stat(note); os.SameFile(was, now) {
		t.Error("a save wrote over the file in place")
	}
	if fi, err := os.Lstat(link); err != nil ||
		fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("a save through a link left no link: %v", err)
	}
	if b, _ := os.ReadFile(note); string(b) != "xa\n" {
		t.Errorf("a save through a link: the file holds %q", b)
	}
	if fi, _ := os.Stat(note); fi.Mode().Perm() != 0o750 {
		t.Errorf("a save made the file %v, not 0750", fi.Mode())
	}
	other := filepath.Join(dir, "other")
	os.Link(note, other)
	keys(on(t, note), "i", "y", "ESC", "s")
	if b, _ := os.ReadFile(other); string(b) != "yxa\n" {
		t.Errorf("a save over a hard link: its other name holds %q",
			b)
	}
	if names, _ := filepath.Glob(filepath.Join(dir, "*")); len(names) !=
		3 {
		t.Errorf("a save left %v", names)
	}
	if hidden, _ := filepath.Glob(filepath.Join(dir, ".*")); len(hidden) >
		0 {
		t.Errorf("a save left %v beside the file", hidden)
	}
}

// In a frame Tab runs l, which turns the frame's tabs, and types nothing;
// ed knows no tabs itself. Outside a frame there are none, and ed says so.
func TestTab(t *testing.T) {
	bin, ran := t.TempDir(), filepath.Join(t.TempDir(), "ran")
	os.WriteFile(filepath.Join(bin, "l"),
		[]byte("#!/bin/sh\ntouch "+ran+"\n"), 0o755)
	path := "PATH=" + bin + ":" + os.Getenv("PATH")
	_, press := ed(t, "abc\n", path, "STUDIO=view")
	if s := keys(press, "TAB"); s.lines[0] != "abc" {
		t.Errorf("Tab typed: %q", s.lines[0])
	}
	ranL := func() bool { _, err := os.Stat(ran); return err == nil }
	if !settled(ranL) {
		t.Error("Tab in a frame ran no l")
	}
	_, press = ed(t, "abc\n", path)
	if s := keys(press, "TAB"); s.said != "no tabs outside a frame" {
		t.Errorf("Tab outside a frame says %q", s.said)
	}
}

// settled is whether ok comes to hold within a few seconds.
func settled(ok func() bool) bool {
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); {
		if ok() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return ok()
}
