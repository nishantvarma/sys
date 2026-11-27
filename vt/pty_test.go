//go:build linux

package vt

import (
	"io"
	"os"
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

// TestPiped: a program whose input and output are set already keeps
// them, as a host's program keeps the pipe it speaks on, and has the pty
// for its error and its controlling terminal, so that what it runs finds
// the pty as its terminal, /dev/tty, as fzy finds one.
func TestPiped(t *testing.T) {
	var out strings.Builder
	cmd := exec.Command("sh", "-c",
		`read x; echo "$x"; echo on the tty >/dev/tty; echo err >&2`)
	cmd.Stdin, cmd.Stdout = strings.NewReader("piped\n"), &out
	pty, err := Start(cmd, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer pty.Close()
	screen, _ := io.ReadAll(pty) // to the pty's end, EIO, once it exits
	cmd.Wait()
	if got := out.String(); got != "piped\n" {
		t.Errorf("the pipe carries %q, not the input given", got)
	}
	if s := string(screen); !strings.Contains(s, "on the tty") ||
		!strings.Contains(s, "err") {
		t.Errorf("the pty shows %q, not /dev/tty's line and the error",
			s)
	}
	all := exec.Command("true")
	all.Stdin, all.Stdout, all.Stderr = os.Stdin, os.Stdout, os.Stderr
	if _, err := Start(all, 80, 24); err == nil {
		t.Error("a program with all three set started on a pty")
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
