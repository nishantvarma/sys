package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestLog: the log keeps a line a place, under the verb that ran on it
// (`sh/c:/one line per place/`): a select, c with no verb, moves a place's
// line to the newest and keeps its verb, so a script run stays run, and
// an act says its own; m leads with where you are, under its own line's
// verb, and walks that line once, so 1 is one back. c and m are rc, and
// the test skips without it.
func TestLog(t *testing.T) {
	if _, err := exec.LookPath("9"); err != nil {
		t.Skip("no 9")
	}
	home, bin, work := t.TempDir(), t.TempDir(), t.TempDir()
	for _, f := range []string{"c", "m"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "sh", f))
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(bin, f), b, 0o755)
	}
	note := filepath.Join(work, "note")
	script := filepath.Join(work, "sync")
	os.WriteFile(note, nil, 0o644)
	os.WriteFile(script, []byte("#!/bin/sh\n"), 0o755)
	// in no tmux, so that cur is the window none's
	env := append(slices.DeleteFunc(os.Environ(), func(v string) bool {
		return strings.HasPrefix(v, "TMUX")
	}), "HOME="+home, "PATH="+bin+":"+os.Getenv("PATH"))
	say := func(args ...string) string {
		t.Helper()
		c := exec.Command(args[0], args[1:]...)
		c.Env, c.Dir = env, work
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", strings.Join(args, " "), err,
				out)
		}
		return string(out)
	}
	log := filepath.Join(home, ".config", "sys", "log")
	logged := func() string {
		b, _ := os.ReadFile(log)
		return string(b)
	}
	say("c", "-v", "run", script)
	say("c", note)
	say("c", script)
	if got, want := logged(), "c "+note+"\nrun "+script+"\n"; got != want {
		t.Errorf("a select after a run logged %q, not %q", got, want)
	}
	if got, want := say("m"), "run "+script+"\nc "+note+"\n"; got != want {
		t.Errorf("m said %q, not %q", got, want)
	}
	if got := say("m", "1"); got != note+"\n" {
		t.Errorf("m 1 said %q, not one back, %s", got, note)
	}
	say("c", "-v", "e", script)
	if got, want := logged(), "c "+note+"\ne "+script+"\n"; got != want {
		t.Errorf("an act after a run logged %q, not %q", got, want)
	}
}
