// Package test lays a frame out on a tmux server of the test's own, as
// studio does, and ends it as fm's q does, by the end of its face: each
// slot's program says its slot and waits for its input to end.
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

// frame is two programs over two shells, the first program the face.
const frame = `a frame: two programs over two shells
top  70%
  a  50%  sh -c 'echo in $STUDIO; exec cat'
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

// TestEnd: each slot's program is told its slot; a shell that exits ends
// alone; the face's end ends the frame, the rest of its programs and the
// shells at their prompt, and the window with them; a shell running a
// command stays, and its window is then a plain one, with no unsaved file
// left for it.
func TestEnd(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux")
	}
	studio := built(t)
	home := t.TempDir()
	sock := fmt.Sprintf("sys-studio-%d", os.Getpid())
	env := append(os.Environ(), "TMUX_PANE=", "HOME="+home)
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

	// open is the frame in a new window, as studio lays it out from a
	// shell that is not alone in its window, and its slots' panes
	open := func() (string, map[string]string) {
		c := exec.Command(studio, note)
		c.Dir, c.Env = t.TempDir(), append(env, "TMUX="+server)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("studio: %v: %s", err, out)
		}
		w := tm("show", "-gv", "@studio")
		slot := map[string]string{}
		for _, n := range []string{"a", "b", "one", "two"} {
			slot[n] = tm("show", "-wv", "-t", w, "@studio-"+n)
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
	says := func(p, s string) func() bool {
		return func() bool {
			out := tm("capture-pane", "-p", "-t", p)
			return strings.Contains(out, s)
		}
	}

	w, slot := open()
	for _, n := range []string{"a", "b"} {
		if !settle(says(slot[n], "in "+n)) {
			t.Errorf("slot %s's program is not told its slot", n)
		}
	}
	tm("send-keys", "-t", slot["a"], "C-d")
	if !settle(func() bool { return len(panes(w)) == 0 }) {
		t.Errorf("the face ended, and %v stay", panes(w))
	}

	w, slot = open()
	settle(says(slot["a"], "in a"))
	for _, n := range []string{"one", "two"} { // typed at its prompt
		settle(func() bool {
			return tm("capture-pane", "-p", "-t", slot[n]) != ""
		})
	}
	tm("send-keys", "-t", slot["one"], "exit", "Enter")
	if !settle(func() bool {
		return !slices.Contains(panes(w), slot["one"])
	}) {
		t.Fatal("a shell's exit did not end it")
	}
	time.Sleep(300 * time.Millisecond) // the hook it ran has done
	if got := panes(w); len(got) != 3 {
		t.Errorf("a shell exited, and the frame is %v", got)
	}
	tm("send-keys", "-t", slot["two"], "sleep 100", "Enter")
	if !settle(func() bool {
		return tm("display", "-p", "-t", slot["two"],
			"#{pane_current_command}") == "sleep"
	}) {
		t.Fatal("the shell runs no command")
	}
	unsaved := filepath.Join(home, ".config", "sys", "unsaved", w)
	os.MkdirAll(filepath.Dir(unsaved), 0o755)
	os.WriteFile(unsaved, []byte("/a/file\n"), 0o644)
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
	if _, err := os.Stat(unsaved); err == nil {
		t.Error("the frame ended, and its unsaved file stays")
	}
}
