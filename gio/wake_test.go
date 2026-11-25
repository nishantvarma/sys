// every reply from a program must wake the window, or it is not drawn until
// something else does
package main

import (
	"image"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// the window as it runs: it draws only when woken or given input
func TestWake(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.txt", "b.txt", "c.txt", "d.txt"} {
		os.WriteFile(dir+"/"+f, []byte(f+"\n"), 0o644)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", programs(t)+":"+os.Getenv("PATH"))
	s, err := start(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, p := range s.panes {
			if p.cmd != nil {
				p.cmd.Process.Kill()
			}
		}
	}()
	v, err := newWindow(s, true)
	if err != nil {
		t.Fatal(err)
	}
	woken := make(chan struct{}, 1)
	var wakes atomic.Int32
	v.listen(func() {
		wakes.Add(1)
		select {
		case woken <- struct{}{}:
		default:
		}
	})
	var router input.Router
	var ops op.Ops
	frame := func() {
		ops.Reset()
		v.draw(layout.Context{Ops: &ops, Source: router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1360, 800))})
		router.Frame(&ops)
	}
	// idle draws each time it is woken, until no wake comes for a while
	idle := func() {
		for {
			select {
			case <-woken:
				frame()
			case <-time.After(500 * time.Millisecond):
				return
			}
		}
	}
	frame()
	frame()
	idle()
	fm, ed := &s.panes["fm"].app, &s.panes["ed"].app
	for _, want := range []string{"b.txt", "c.txt", "d.txt"} {
		before := wakes.Load()
		router.Queue(key.EditEvent{Text: "j"})
		frame() // the key's own frame, and no more unless woken
		idle()
		got := cursor(fm)
		t.Logf(
			"j: %d wakes, fm shows %q, ed shows %q, "+
				"news unseen %d",
			wakes.Load()-before,
			got,
			head(ed)[len(head(ed))-min(5, len(head(ed))):],
			len(v.news),
		)
		if got != want || head(ed) != dir+"/"+want {
			t.Errorf("after j: want %s on screen in both", want)
		}
	}
}
