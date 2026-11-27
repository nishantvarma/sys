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
