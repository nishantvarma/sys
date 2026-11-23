package outline

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// An address goes out and comes back whole, a buffer's with no file too;
// what has no line above 0 after its last colon is all path, line 0.
func TestAddress(t *testing.T) {
	if a := At("/a/b", 12); a != "/a/b:12" {
		t.Errorf("At: %s, want /a/b:12", a)
	}
	for _, c := range []struct {
		id, path string
		line     int
	}{
		{"/a/b:12", "/a/b", 12},
		{"/a/b", "/a/b", 0},
		{"/a/b:", "/a/b:", 0},
		{"/a/b:x", "/a/b:x", 0},
		{"/a/b:0", "/a/b:0", 0},
		{"/a:b/c:3", "/a:b/c", 3},
		{":12", "", 12},
	} {
		path, line := Address(c.id)
		if path != c.path || line != c.line {
			t.Errorf("%q: %q %d, want %q %d", c.id, path, line,
				c.path, c.line)
		}
	}
}

// File takes text a person types and refuses the rest, saying why: a
// folder, a fifo, which it never opens, a binary, and a file over a
// megabyte.
func TestFile(t *testing.T) {
	d := t.TempDir()
	write := func(name string, b []byte) string {
		p := filepath.Join(d, name)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	note := write("note", []byte("a\n  b\n"))
	if b, err := File(note); err != nil || string(b) != "a\n  b\n" {
		t.Errorf("a note: %q %v", b, err)
	}
	fifo := filepath.Join(d, "fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		d,
		fifo,
		write("bin", []byte("a\x00b")),
		write("big", []byte(strings.Repeat("a", 1<<20+1))),
	} {
		if _, err := File(p); err == nil {
			t.Errorf("%s: taken as a note", p)
		}
	}
}

// A note's head is its first item, and the items under it as one line.
func TestHead(t *testing.T) {
	items := Text(strings.NewReader("t\n  one\n    two\nx\n"), "n")
	title, lead := Head(items)
	if title != "t" || lead != "one two" {
		t.Errorf("head: %q %q, want t, one two", title, lead)
	}
}
