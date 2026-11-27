package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestButtons: on a client of the test's own, tmux's conf sourced, a press
// on a name on a tab line shows that tab, the panel's + opens a terminal,
// and its × closes the one shown, each a press where the name is drawn.
// The client's presses go in as a terminal sends them; no sys script is on
// the server's path and it has no display, so no hook moves a pointer.
func TestButtons(t *testing.T) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("no tmux")
	}
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("no script, to give a client a terminal")
	}
	studio := built(t)
	home, bin := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(bin, "fm"),
		[]byte("#!/bin/sh\nexec cat\n"), 0o755)
	// the server's programs run this tmux, as the test does: another on
	// /usr/bin speaks another version of tmux's protocol
	os.Symlink(tmux, filepath.Join(bin, "tmux"))
	path := filepath.Dir(studio) + ":" + bin + ":/usr/bin:/bin"
	sock := fmt.Sprintf("sys-buttons-%d", os.Getpid())
	env := append(os.Environ(), "TMUX_PANE=", "HOME="+home, "DISPLAY=",
		"PS1=$ ", "TERM=xterm-256color", "PATH="+path)
	tm := func(args ...string) (string, error) {
		c := exec.Command("tmux", append([]string{"-L", sock, "-f",
			"/dev/null"}, args...)...)
		c.Env = append(env, "TMUX=")
		out, err := c.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	must := func(args ...string) string {
		t.Helper()
		out, err := tm(args...)
		if err != nil {
			t.Fatalf("tmux %s: %v: %s",
				strings.Join(args, " "), err, out)
		}
		return out
	}
	must("new-session", "-d", "-s", "main", "-x", "200", "-y", "50",
		"sleep 100")
	t.Cleanup(func() { tm("kill-server") })
	if out, err := tm("bind", "-n", "MouseDown1Control0",
		"select-pane"); err != nil {
		t.Skipf("tmux has no control ranges: %s", out)
	}
	must("set-environment", "-g", "PATH", path)
	must("source-file", "../../tmux/conf")
	server := must("display", "-p", "#{socket_path},#{pid},0")
	note := filepath.Join(t.TempDir(), "frame")
	os.WriteFile(note, []byte(tabbed), 0o644)
	run := func(pane string, args ...string) {
		t.Helper()
		c := exec.Command(studio, args...)
		c.Dir, c.Env = t.TempDir(), append(env, "TMUX="+server,
			"TMUX_PANE="+pane)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("studio %s: %v: %s", strings.Join(args, " "),
				err, out)
		}
	}
	run("", note)
	w := must("show", "-gv", "@studio")
	in := func(slot string) (string, string) {
		out := must("list-panes", "-t", w, "-f", "#{==:#{@slot},"+slot+
			"}", "-F", "#{pane_id} #{@tab}")
		p, tab, _ := strings.Cut(out, " ")
		return p, tab
	}
	shell, _ := in("shell")
	run(shell, "tab", "new")

	client := exec.Command("script", "-qfc",
		"stty rows 50 cols 200; exec tmux -L "+sock+" attach -t main",
		"/dev/null")
	client.Env = append(env, "TMUX=")
	keys, err := client.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	client.Stdout = io.Discard
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		keys.Close()
		client.Process.Kill()
		client.Wait()
	})
	if !settle(func() bool {
		out, _ := tm("list-clients", "-F", "#{client_name}")
		return out != ""
	}) {
		t.Fatal("no client attached")
	}
	// the window takes the client's size
	time.Sleep(300 * time.Millisecond)

	// press is a left click at a cell, 0-based, as a terminal sends one
	press := func(x, y int) {
		fmt.Fprintf(keys, "\x1b[<0;%d;%dM", x+1, y+1)
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintf(keys, "\x1b[<0;%d;%dm", x+1, y+1)
	}
	styles := regexp.MustCompile(`#\[[^\]]*\]`)
	// on is where a label is drawn on the tab line over a slot's pane:
	// the line over the pane, the format two cells in
	on := func(slot, label string) (int, int) {
		t.Helper()
		var p, said string
		i := -1
		settle(func() bool { // a tab shown is named on its line after
			p, _ = in(slot)
			bar, _ := tm("show", "-pqv", "-t", p, "@bar")
			said = styles.ReplaceAllString(bar, "")
			i = strings.Index(said, " "+label+" ")
			return i >= 0
		})
		at := strings.Fields(must("display", "-p", "-t", p,
			"#{pane_left} #{pane_top}"))
		if i < 0 || len(at) != 2 {
			t.Fatalf("no %q on %s's tab line: %q", label, slot,
				said)
		}
		left, _ := strconv.Atoi(at[0])
		top, _ := strconv.Atoi(at[1])
		return left + 2 + i + 1, top - 1
	}
	shows := func(slot, tab string) bool {
		return settle(func() bool {
			_, got := in(slot)
			return got == tab
		})
	}

	press(on("view", "second"))
	if !shows("view", "second") {
		_, got := in("view")
		t.Errorf("a press on second, and the view shows %q", got)
	}
	press(on("shell", "1"))
	if !shows("shell", "shell") {
		_, got := in("shell")
		t.Errorf("a press on 1, and the panel shows %q", got)
	}
	press(on("shell", "+"))
	if !shows("shell", "3") {
		_, got := in("shell")
		t.Errorf("a press on +, and the panel shows %q", got)
	}
	press(on("shell", "×"))
	if !settle(func() bool {
		out, _ := tm("list-panes", "-a", "-f", "#{&&:#{==:#{@frame},"+
			w+"},#{==:#{@slot},shell}}", "-F", "#{@tab}")
		return !strings.Contains(out, "3")
	}) {
		t.Error("a press on ×, and terminal 3 stays")
	}
}
