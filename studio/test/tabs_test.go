package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// tabbed is a frame with a slot of two tabs, stand-ins that say their
// slot, the face, and a shell.
const tabbed = `a frame: a slot of tabs, the face and a shell
view
  first   sh -c 'echo in $STUDIO first; exec cat'
  second  sh -c 'echo in $STUDIO second; exec cat'
fm        fm
shell
`

// TestTabs: a slot of tabs shows its first in its place and parks the
// rest out of sight, each told its slot, and the border over the slot
// names them; next shows the second there, parking the first; new opens
// a terminal as a tab of the panel, the frame's shell, told no slot, its
// bar numbering the terminals, and a number shows the terminal at that
// place; the panel hides, its room going to the slots above, and shows
// again across the bottom, as a new terminal shows it too; the terminal's
// exit shows the shell again, and so does its close from its own shell,
// the tab line naming it no more; a press's error is answered, not
// failed; close ends the tab shown, the next taking its place; and the
// face's end ends the parked tabs with the frame.
func TestTabs(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("no tmux")
	}
	studio := built(t)
	home, bin := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(bin, "fm"),
		[]byte("#!/bin/sh\necho in $STUDIO\nexec cat\n"), 0o755)
	sock := fmt.Sprintf("sys-tabs-%d", os.Getpid())
	env := append(os.Environ(), "TMUX_PANE=", "HOME="+home,
		"PROMPT_COMMAND=", "PS1=$ ", "PATH="+bin+":"+os.Getenv("PATH"))
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
	os.WriteFile(note, []byte(tabbed), 0o644)
	// studio runs studio's verb from the pane given, as a slot's program
	// runs it
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
	w := tm("show", "-gv", "@studio")
	// in is the pane window w shows in a slot, and the tab it holds
	in := func(slot string) (string, string) {
		out := tm("list-panes", "-t", w, "-f", "#{==:#{@slot},"+slot+
			"}", "-F", "#{pane_id} #{@tab}")
		p, tab, _ := strings.Cut(out, " ")
		return p, tab
	}
	// tabs is a slot's tabs, wherever each waits
	tabs := func(slot string) string {
		return tm("list-panes", "-a", "-f", "#{&&:#{==:#{@frame},"+w+
			"},#{==:#{@slot},"+slot+"}}", "-F", "#{@tab}")
	}
	screen := func(p string) string {
		return tm("capture-pane", "-p", "-t", p)
	}

	view, tab := in("view")
	if tab != "first" || !settle(func() bool {
		return strings.Contains(screen(view), "in view first")
	}) {
		t.Fatalf("the slot shows %q: %q", tab, screen(view))
	}
	if got := tabs("view"); got != "first\nsecond" {
		t.Errorf("the slot's tabs: %q", got)
	}
	bar := tm("show", "-pqv", "-t", view, "@bar")
	if !strings.Contains(bar, " first ") ||
		!strings.Contains(bar, " second ") ||
		tm("show", "-wqv", "-t", w, "pane-border-status") != "top" {
		t.Errorf("the border names no tabs: %q", bar)
	}

	run(view, "tab", "next")
	second, tab := in("view")
	if tab != "second" || !settle(func() bool {
		return strings.Contains(screen(second), "in view second")
	}) {
		t.Errorf("next shows %q: %q", tab, screen(second))
	}

	shell, _ := in("shell")
	run(shell, "tab", "new")
	term, tab := in("shell")
	if tab != "2" || term == shell {
		t.Fatalf("new shows %q in the shell's place", tab)
	}
	tm("send-keys", "-t", term, "echo told $STUDIO.", "Enter")
	if !settle(func() bool {
		return strings.Contains(screen(term), "\ntold .")
	}) {
		t.Errorf("a new terminal is told a slot: %q", screen(term))
	}
	if bar := tm("show", "-pqv", "-t", term, "@bar"); !strings.Contains(
		bar, " 1 ") || !strings.Contains(bar, " 2 ") {
		t.Errorf("the panel's bar numbers no terminals: %q", bar)
	}
	run(term, "tab", "1")
	if p, _ := in("shell"); p != shell {
		t.Errorf("tab 1 shows %s, not the first terminal", p)
	}
	panes := func() string {
		return tm("display", "-p", "-t", w, "#{window_panes}")
	}
	run(shell, "panel", "hide")
	if p, _ := in("shell"); p != "" || panes() != "2" {
		t.Errorf("the panel hid, and the frame is %s panes", panes())
	}
	run("", "panel", "-t", view) // by a pane, as alt-t's binding asks
	if p, _ := in("shell"); p != shell || tm("display", "-p", "-t", p,
		"#{pane_width}") != tm("display", "-p", "-t", w,
		"#{window_width}") {
		t.Errorf("the panel showed %q, not the shell across the"+
			" bottom", p)
	}
	run(shell, "panel", "hide")
	run(view, "tab", "2")
	if p, _ := in("shell"); p != "" {
		t.Errorf("tab 2 of the view showed the panel: %s", p)
	}
	run(view, "tab", "-t", term, "1") // by the terminal's pane
	if p, _ := in("shell"); p != shell {
		t.Errorf("a terminal's tab shown, and the panel shows %q", p)
	}
	run(shell, "tab", "2")
	tm("send-keys", "-t", term, "exit", "Enter")
	if !settle(func() bool {
		_, tab := in("shell")
		return tab == "shell" && tabs("shell") == "shell"
	}) {
		_, tab := in("shell")
		t.Errorf("the terminal exited, and the shell slot shows %q,"+
			" its tabs %q", tab, tabs("shell"))
	}
	run(shell, "tab", "new")
	term, _ = in("shell")
	tm("send-keys", "-t", term, studio+" tab close", "Enter")
	bar = ""
	if !settle(func() bool {
		p, tab := in("shell")
		bar = tm("show", "-pqv", "-t", p, "@bar")
		return tab == "shell" && tabs("shell") == "shell" &&
			!strings.Contains(bar, " 2 ")
	}) {
		t.Errorf("a terminal closed from its own shell, and the tab"+
			" line says %q", bar)
	}
	run("", "tab", "-t", shell, "9") // a press: answered, not failed

	run(second, "tab", "close")
	if _, tab := in("view"); tab != "first" || tabs("view") != "first" {
		t.Errorf("close shows %q, the tabs %q", tab, tabs("view"))
	}

	fm, _ := in("fm")
	tm("send-keys", "-t", fm, "C-d")
	if !settle(func() bool {
		back := "=studio-" + strings.TrimPrefix(w, "@")
		c := exec.Command("tmux", "-L", sock, "-f", "/dev/null",
			"has-session", "-t", back)
		c.Env = append(env, "TMUX=")
		return c.Run() != nil
	}) {
		t.Error("the face ended, and the parked tabs stay")
	}
}
