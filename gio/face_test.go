// the frame's end: by its face's, fm's
package main

import (
	"os"
	"path/filepath"
	"testing"

	"sys/state"
)

// TestFace: fm's end ends the frame, but not while ed names work it holds
// unsaved, which fm's second q takes away: an fm that ended some other
// way, as a crash ends it, leaves the frame for ed to save. Another's end
// is its own, the eye moving on from it.
func TestFace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	unsaved := state.Path("unsaved", "none")
	frame := func() (*studio, *pane, *pane) {
		fm, ed := &pane{name: "fm"}, &pane{name: "ed"}
		return &studio{order: []*pane{fm, ed}, face: fm,
			unsaved: unsaved}, fm, ed
	}
	s, _, ed := frame()
	s.eye = 1
	if _, done := s.handle(news{p: ed, ended: true}); done || s.eye != 0 {
		t.Errorf("ed ended: the window done %v, the eye on %d", done,
			s.eye)
	}
	os.MkdirAll(filepath.Dir(unsaved), 0o755)
	os.WriteFile(unsaved, []byte("/a.txt\n"), 0o644)
	s, fm, _ := frame()
	if _, done := s.handle(news{p: fm, ended: true}); done {
		t.Error("fm ended over unsaved work, and the window went")
	}
	os.Remove(unsaved)
	s, fm, _ = frame()
	if _, done := s.handle(news{p: fm, ended: true}); !done {
		t.Error("fm ended, and the window stays")
	}
}
