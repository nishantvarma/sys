package state

import (
	"os"
	"path/filepath"
	"testing"
)

// A line written is the line read, whole, from the place `sh/c` keeps it;
// a file not written yet reads as nothing.
func TestState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := Path("cur", "@3")
	if want := filepath.Join(home, ".config/sys/cur/@3"); p != want {
		t.Errorf("path: %s, want %s", p, want)
	}
	if got := Read(p); got != "" {
		t.Errorf("unwritten: %q", got)
	}
	if err := Write(p, "/a/b:12"); err != nil {
		t.Fatal(err)
	}
	if got := Read(p); got != "/a/b:12" {
		t.Errorf("read: %q, want /a/b:12", got)
	}
	left, err := filepath.Glob(p + ".*")
	if err != nil || len(left) > 0 {
		t.Errorf("a temporary file is left: %v %v", left, err)
	}
	if b, _ := os.ReadFile(p); string(b) != "/a/b:12\n" {
		t.Errorf("file: %q", b)
	}
	// ed's text is put whole, every line of it
	text := Path("text", "@3")
	if err := Put(text, []byte("a\n  b\n")); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(text); string(b) != "a\n  b\n" {
		t.Errorf("text: %q", b)
	}
}

// A slot's program reads its slot once, and what it starts reads none.
func TestSlot(t *testing.T) {
	t.Setenv("STUDIO", "ed")
	if s := Slot(); s != "ed" {
		t.Errorf("slot: %q, want ed", s)
	}
	if s, ok := os.LookupEnv("STUDIO"); ok {
		t.Errorf("STUDIO is still %q", s)
	}
}

// Outside tmux the window is none and the eye is here.
func TestPane(t *testing.T) {
	t.Setenv("TMUX_PANE", "")
	if w, eye := Pane(); w != "none" || !eye {
		t.Errorf("outside tmux: %s %v, want none true", w, eye)
	}
}
