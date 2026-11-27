//go:build linux

package vt

import (
	"io"
	"os/exec"
	"strings"
	"testing"
)

// TestStart: a program run on a pty sees st's TERM and the size given,
// and a new one once told; what it writes, a terminal puts on its
// screen.
func TestStart(t *testing.T) {
	cmd := exec.Command("sh", "-c",
		`printf '%s %s;' "$TERM" "$(stty size)"; read x; stty size`)
	pty, err := Start(cmd, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer pty.Close()
	buf := make([]byte, 64)
	n, _ := pty.Read(buf)
	if got := string(buf[:n]); got != "st-256color 24 80;" {
		t.Errorf("the program sees %q", got)
	}
	if err := Setsize(pty, 100, 30); err != nil {
		t.Fatal(err)
	}
	pty.Write([]byte("\n"))
	out, _ := io.ReadAll(pty) // to the pty's end, EIO, once it exits
	cmd.Wait()
	if !strings.Contains(string(out), "30 100") {
		t.Errorf("told 100 by 30, the program sees %q", out)
	}
	tm := New(20, 3, io.Discard)
	tm.Write(out)
	if got := lines(tm); !strings.Contains(strings.Join(got, "|"),
		"30 100") {
		t.Errorf("the screen holds %q", got)
	}
}

// TestForeground: the program in a pty's foreground is the one it runs,
// its session's leader, until that one runs another in a group of its
// own; and a program started with a folder and no world of its own is
// told the folder, PWD, as os/exec tells it.
func TestForeground(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("sh", "-c", `echo "$PWD"; read x`)
	cmd.Dir = dir
	pty, err := Start(cmd, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		pty.Close()
		cmd.Wait()
	}()
	if g, err := Foreground(pty); err != nil || g != cmd.Process.Pid {
		t.Errorf("the foreground is %d, %v, not the program's %d", g,
			err, cmd.Process.Pid)
	}
	buf := make([]byte, 256)
	n, _ := pty.Read(buf)
	if got := strings.TrimSpace(string(buf[:n])); got != dir {
		t.Errorf("the program's PWD is %q, not %q", got, dir)
	}
}
