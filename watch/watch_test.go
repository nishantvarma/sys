package watch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// heard is whether w says a change after write: what is still being said
// of what came before is let pass first, for one change can be several
// words.
func heard(t *testing.T, w *Watcher, write func() error) bool {
	t.Helper()
	for quiet := false; !quiet; {
		select {
		case <-w.On:
		case <-time.After(50 * time.Millisecond):
			quiet = true
		}
	}
	if err := write(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.On:
		return true
	case <-time.After(300 * time.Millisecond):
		return false
	}
}

// A file renamed into a watched folder, as fm writes cur, is heard; so is
// one written in a followed folder; a write elsewhere is not.
func TestHeard(t *testing.T) {
	fixed, followed, other := t.TempDir(), t.TempDir(), t.TempDir()
	w, err := New(fixed)
	if err != nil {
		t.Fatal(err)
	}
	w.Follow(followed)
	rename := func() error {
		tmp := filepath.Join(fixed, "cur.1")
		if err := os.WriteFile(tmp, []byte("a\n"), 0o644); err != nil {
			return err
		}
		return os.Rename(tmp, filepath.Join(fixed, "cur"))
	}
	if !heard(t, w, rename) {
		t.Error("a rename into the fixed folder was not heard")
	}
	save := func() error {
		return os.WriteFile(filepath.Join(followed, "n"), nil, 0o644)
	}
	if !heard(t, w, save) {
		t.Error("a save in the followed folder was not heard")
	}
	elsewhere := func() error {
		return os.WriteFile(filepath.Join(other, "x"), nil, 0o644)
	}
	if heard(t, w, elsewhere) {
		t.Error("a write in a folder not watched was heard")
	}
}

// Following a fixed folder by another name, a link to it, and then moving
// on keeps its watch: a change there is still heard.
func TestFixedKept(t *testing.T) {
	fixed, away := t.TempDir(), t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(fixed, link); err != nil {
		t.Fatal(err)
	}
	w, err := New(fixed)
	if err != nil {
		t.Fatal(err)
	}
	w.Follow(link)
	w.Follow(away)
	write := func() error {
		return os.WriteFile(filepath.Join(fixed, "cur"), nil, 0o644)
	}
	if !heard(t, w, write) {
		t.Error("the fixed folder's watch was let go")
	}
}

// A folder that cannot be watched is an error that names it.
func TestNamed(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	_, err := New(gone)
	if err == nil || !strings.Contains(err.Error(), gone) {
		t.Errorf("error: %v, want one naming %s", err, gone)
	}
}
