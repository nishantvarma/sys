// Package tmux holds sys's tmux layer, conf, and its check: what the
// layout hooks do, run on a tmux server of the test's own.
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// server is a tmux server of the test's own, conf sourced, and a way to
// run its commands: one answer, trimmed. Its programs find no sys on
// their path, so a hook that warps the pointer warps nothing.
func server(t *testing.T) func(args ...string) string {
	bin, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("no tmux")
	}
	conf, err := filepath.Abs("conf")
	if err != nil {
		t.Fatal(err)
	}
	sock := fmt.Sprintf("sys-test-%d", os.Getpid())
	tm := func(args ...string) string {
		t.Helper()
		c := exec.Command("tmux", append([]string{"-L", sock, "-f",
			"/dev/null"}, args...)...)
		c.Env = append(os.Environ(), "TMUX=", "TMUX_PANE=",
			"PATH="+filepath.Dir(bin)+":/usr/bin:/bin")
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("tmux %s: %v: %s", strings.Join(args, " "),
				err, out)
		}
		return strings.TrimSpace(string(out))
	}
	tm("new-session", "-d", "-x", "200", "-y", "50", "sleep 100")
	t.Cleanup(func() { tm("kill-server") })
	tm("source-file", conf)
	return tm
}

// settle waits, up to a second, until ok holds.
func settle(ok func() bool) bool {
	for end := time.Now().Add(time.Second); time.Now().Before(end); {
		if ok() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return ok()
}

// TestLayout: a window is laid out again, main-vertical, when its own
// panes change in number, a pane ending or killed in it, and no other
// window is: a frame laid out by hand, its hook emptied as studio empties
// it, keeps its layout as panes end and windows close elsewhere, and so
// does a window resized by hand.
func TestLayout(t *testing.T) {
	tm := server(t)
	layout := func(w string) string {
		return tm("display", "-p", "-t", w, "#{window_layout}")
	}
	main := func(w string) func() bool { // main-vertical: side by side
		return func() bool { return strings.Contains(layout(w), "{") }
	}
	window := func(args ...string) string {
		return tm(append([]string{"new-window", "-P", "-F",
			"#{window_id}"}, args...)...)
	}
	frame := window("sleep 100")
	tm("set-hook", "-w", "-t", frame, "window-layout-changed", "")
	tm("split-window", "-d", "-v", "-t", frame, "-l", "30%", "sleep 100")
	tm("split-window", "-d", "-h", "-t", frame+".0", "-l", "60%",
		"sleep 100")
	framed := layout(frame)

	// a window of three panes stacked, the last ending after a while
	// unless it is killed first; d, a window not looked at
	stacked := func(d bool) (string, string) {
		o := window("sleep 100")
		if d {
			tm("select-window", "-t", frame)
		}
		tm("split-window", "-d", "-t", o, "sleep 100")
		p := tm("split-window", "-d", "-P", "-F", "#{pane_id}", "-t",
			o, "sleep 0.3")
		tm("select-layout", "-t", o, "even-vertical")
		return o, p
	}
	o, _ := stacked(true)
	if !settle(main(o)) {
		t.Errorf("a pane ended in a window not looked at: %s, not"+
			" main-vertical", layout(o))
	}
	o, p := stacked(true)
	tm("kill-pane", "-t", p)
	if !settle(main(o)) {
		t.Errorf("a pane killed in a window not looked at: %s, not"+
			" main-vertical", layout(o))
	}
	o, _ = stacked(false)
	if !settle(main(o)) {
		t.Errorf("a pane ended in the window looked at: %s, not"+
			" main-vertical", layout(o))
	}
	tm("resize-pane", "-t", o+".1", "-L", "30")
	resized := layout(o)
	tm("select-window", "-t", frame)
	// a window ends whole, then a pane ends in another: once that other
	// is laid out again both hooks have run, and neither may have touched
	// the frame or the window resized by hand
	window("-d", "sleep 0.1")
	last, _ := stacked(true)
	if !settle(main(last)) {
		t.Fatalf("a pane ended in a window not looked at: %s, not"+
			" main-vertical", layout(last))
	}
	if got := layout(frame); got != framed {
		t.Errorf("the frame went from %s to %s", framed, got)
	}
	if got := layout(o); got != resized {
		t.Errorf("a window resized by hand went from %s to %s",
			resized, got)
	}
}
