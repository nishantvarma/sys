// Package test lays a frame out on a tmux server of the test's own, as
// studio does, and ends it as fm's q does, by the end of its face: each
// slot's program says its slot and waits for its input to end, the face's
// a stand-in for fm.
package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// built is studio made from this tree, as ./build makes it: tangled, then
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
	run("..", "go", "build", "-o", filepath.Join(dir, "studio"),
		"main.go")
	return filepath.Join(dir, "studio")
}

// frame is fm, the face, and another program over two shells.
const frame = `a frame: two programs over two shells
top  70%
  a  50%  fm
  b       sh -c 'echo in $STUDIO; exec cat'
bottom
  one
  two
`

// settle waits, up to two seconds, until ok holds.
func settle(ok func() bool) bool {
	for end := time.Now().Add(2 * time.Second); time.Now().Before(end); {
		if ok() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return ok()
}

// TestEnd: each slot's program is told its slot; the face's end ends the
// frame, the other program and the shells at their prompt, and the window
// with them; a shell that exits ends alone, and so does a face that ends
// while ed names work it holds unsaved, as after a crash; and a shell
// running a command stays, its window then a plain one.
func TestEnd(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux")
	}
	studio := built(t)
	home, bin := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(bin, "fm"),
		[]byte("#!/bin/sh\necho in $STUDIO\nexec cat\n"), 0o755)
	sock := fmt.Sprintf("sys-studio-%d", os.Getpid())
	env := append(os.Environ(), "TMUX_PANE=", "HOME="+home,
		"PATH="+bin+":"+os.Getenv("PATH"))
	tm := func(args ...string) string {
		t.Helper()
		c := exec.Command("tmux", append([]string{"-L", sock, "-f",
			"/dev/null"}, args...)...)
		c.Env = append(env, "TMUX=")
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("tmux %s: %v: %s", strings.Join(args, " "),
				err, out)
		}
		return strings.TrimSpace(string(out))
	}
	tm("new-session", "-d", "-x", "200", "-y", "50", "sleep 100")
	t.Cleanup(func() { tm("kill-server") })
	server := tm("display", "-p", "#{socket_path},#{pid},0")
	note := filepath.Join(t.TempDir(), "frame")
	os.WriteFile(note, []byte(frame), 0o644)

	// shows is whether a pane's screen holds s; for s empty, anything
	shows := func(p, s string) func() bool {
		return func() bool {
			out := tm("capture-pane", "-p", "-t", p)
			return out != "" && strings.Contains(out, s)
		}
	}
	// open is the frame in a new window, as studio lays it out from a
	// shell not alone in its window, and its slots' panes, once each
	// has started: a program saying its slot, a shell its prompt
	open := func() (string, map[string]string) {
		c := exec.Command(studio, note)
		c.Dir, c.Env = t.TempDir(), append(env, "TMUX="+server)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("studio: %v: %s", err, out)
		}
		w := tm("show", "-gv", "@studio")
		slot := map[string]string{}
		for n, says := range map[string]string{"a": "in a",
			"b": "in b", "one": "", "two": ""} {
			slot[n] = tm("show", "-wv", "-t", w, "@studio-"+n)
			if !settle(shows(slot[n], says)) {
				t.Fatalf("slot %s never showed %q", n, says)
			}
		}
		return w, slot
	}
	panes := func(w string) []string { // none once the window is gone
		var in []string
		for _, l := range strings.Split(tm("list-panes", "-a", "-F",
			"#{window_id} #{pane_id}"), "\n") {
			if win, p, _ := strings.Cut(l, " "); win == w {
				in = append(in, p)
			}
		}
		return in
	}
	// keeps is whether the window comes to n panes and keeps them once
	// the hook a pane's end ran has had its time
	keeps := func(w string, n int) bool {
		if !settle(func() bool { return len(panes(w)) == n }) {
			return false
		}
		time.Sleep(300 * time.Millisecond)
		return len(panes(w)) == n
	}

	w, slot := open()
	tm("send-keys", "-t", slot["a"], "C-d")
	if !settle(func() bool { return len(panes(w)) == 0 }) {
		t.Errorf("the face ended, and %v stay", panes(w))
	}

	w, slot = open()
	tm("send-keys", "-t", slot["one"], "exit", "Enter")
	if !keeps(w, 3) {
		t.Errorf("a shell exited, and the frame is %v", panes(w))
	}
	unsaved := filepath.Join(home, ".config", "sys", "unsaved", w)
	os.MkdirAll(filepath.Dir(unsaved), 0o755)
	os.WriteFile(unsaved, []byte("/a/file\n"), 0o644)
	tm("send-keys", "-t", slot["a"], "C-d")
	if !keeps(w, 2) {
		t.Errorf("the face ended while ed held work unsaved, and "+
			"the frame is %v", panes(w))
	}

	w, slot = open()
	tm("send-keys", "-t", slot["two"], "sleep 100", "Enter")
	if !settle(func() bool {
		return tm("display", "-p", "-t", slot["two"],
			"#{pane_current_command}") == "sleep"
	}) {
		t.Fatal("the shell runs no command")
	}
	tm("send-keys", "-t", slot["a"], "C-d")
	if !settle(func() bool {
		return slices.Equal(panes(w), []string{slot["two"]})
	}) {
		t.Errorf("the face ended over a busy shell: %v", panes(w))
	}
	if !settle(func() bool {
		return tm("show", "-wqv", "-t", w, "@studio-frame") == ""
	}) {
		t.Error("the window left is still the frame's")
	}
}
