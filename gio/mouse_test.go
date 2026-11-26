// the mouse in the window: real ed -s and fm -s at the other end of the
// pipe, and studio's frame, presses through gio's own router
package main

import (
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"sys/outline"
)

// alone is a window with one program at the other end of its pipe, run
// from bin in env, and a way to draw until it has been quiet a while
func alone(t *testing.T, env []string, args ...string) (*window,
	*input.Router, func()) {
	c := exec.Command(args[0], args[1:]...)
	c.Env = env
	to, _ := c.StdinPipe()
	from, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Process.Kill(); c.Wait() })
	v, err := newWindow(piped(from, to), false)
	if err != nil {
		t.Fatal(err)
	}
	router := &input.Router{}
	var ops op.Ops
	one := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	return v, router, func() {
		for quiet := 0; quiet < 10; quiet++ {
			if len(v.news) > 0 {
				quiet = 0
			}
			ops.Reset()
			v.draw(layout.Context{Ops: &ops,
				Source:      router.Source(),
				Metric:      one,
				Constraints: layout.Exact(image.Pt(760, 560))})
			router.Frame(&ops)
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// at is the point in the middle of the cell at column x, row y of a pane
// whose corner is c, in a window of cells cw wide
func at(c image.Point, cw, x, y int) f32.Point {
	return f32.Pt(float32(c.X+x*cw+cw/2), float32(c.Y+y*cellH+cellH/2))
}

// TestMouse: in ed's text the wheel moves the cursor three lines, a drag
// selects, a right press hands the word to p, and a middle one the
// selection it falls in to p -a; in fm a right press is enter on the row.
func TestMouse(t *testing.T) {
	bin, said := programs(t), filepath.Join(t.TempDir(), "said")
	stub := t.TempDir() // a p that says what it was handed
	os.WriteFile(filepath.Join(stub, "p"), []byte("#!/bin/sh\n"+
		"{ echo \"$*\"; cat; echo; } >> "+said+"\n"), 0o755)
	env := []string{"HOME=" + t.TempDir(),
		"PATH=" + stub + ":" + bin + ":" + os.Getenv("PATH")}
	t.Setenv("PATH", stub+":"+os.Getenv("PATH"))
	note := filepath.Join(t.TempDir(), "note")
	os.WriteFile(note, []byte("alpha beta\ngamma\ndelta\nepsilon\nzeta\n"),
		0o644)
	v, router, settle := alone(t, env, bin+"/ed", "-s", note)
	settle()
	p, cw := v.s.order[0], v.text.cell
	// ed's text starts under the bar and its rule, after its gutter
	g, top := gutter(len(p.rows)), p.tops()
	word := func(r int) f32.Point {
		return at(image.Point{}, cw, g+r-1, top)
	}
	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse,
		Position: word(1), Scroll: f32.Pt(0, 1)})
	settle()
	if p.cur != 3 {
		t.Errorf("the wheel moved ed to line %d, not 4", p.cur+1)
	}
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse,
		Buttons: pointer.ButtonPrimary, Position: word(1)})
	// a move with the button held, which the router hands on as a drag
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse,
		Buttons: pointer.ButtonPrimary, Position: word(5)})
	router.Queue(pointer.Event{Kind: pointer.Release,
		Source: pointer.Mouse, Position: word(5)})
	settle()
	if got := outline.Selected(p.tree); got != "alpha" {
		t.Errorf("a drag over alpha selected %q", got)
	}
	for _, b := range []pointer.Buttons{pointer.ButtonSecondary,
		pointer.ButtonTertiary} {
		pos := word(8)
		if b == pointer.ButtonTertiary {
			pos = word(3)
		}
		router.Queue(pointer.Event{Kind: pointer.Press,
			Source: pointer.Mouse, Buttons: b, Position: pos})
		router.Queue(pointer.Event{Kind: pointer.Release,
			Source: pointer.Mouse, Position: pos})
		settle()
	}
	want := "-\nbeta\n-a -\nalpha\n"
	for end := time.Now().Add(3 * time.Second); ; {
		b, _ := os.ReadFile(said)
		if string(b) == want {
			break
		}
		if time.Now().After(end) {
			t.Fatalf("p was handed %q, not %q", b, want)
		}
		time.Sleep(20 * time.Millisecond)
	}

	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "alpha"), 0o755)
	os.Mkdir(filepath.Join(dir, "beta"), 0o755)
	v, router, settle = alone(t, env, bin+"/fm", "-s", dir)
	settle()
	fm := v.s.order[0]
	// fm's rows start five lines down, under its head and layouts
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse,
		Buttons:  pointer.ButtonSecondary,
		Position: at(image.Point{}, cw, 6, 5)})
	settle()
	if head(&fm.app) != filepath.Join(dir, "alpha") {
		t.Errorf("a right press on alpha left fm in %q", head(&fm.app))
	}
}

// TestLook: in studio's frame a right press on a file's name in ed's text
// says it, so ed and fm, which follow cur, go there, and the eye with
// them, to ed; nothing opens beside.
func TestLook(t *testing.T) {
	if _, err := exec.LookPath("9"); err != nil {
		t.Skip("no 9, for sys's rc scripts")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("see b.txt\n"),
		0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("beta\n"), 0o644)
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
	var router input.Router
	var ops op.Ops
	one := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	whole := layout.Exact(image.Pt(1500, 800))
	settle := func() {
		for quiet := 0; quiet < 15; quiet++ {
			if len(v.news) > 0 {
				quiet = 0
			}
			ops.Reset()
			v.draw(layout.Context{Ops: &ops,
				Source:      router.Source(),
				Metric:      one,
				Constraints: whole})
			router.Frame(&ops)
			time.Sleep(20 * time.Millisecond)
		}
	}
	ed := s.panes["ed"]
	var corner image.Point
	for end := time.Now().Add(5 * time.Second); ; {
		settle()
		for _, h := range v.hits {
			if h.p == ed {
				corner = h.area.Min
			}
		}
		if strings.HasSuffix(head(&ed.app), "/a.txt") && ed.isText() {
			break
		}
		if time.Now().After(end) {
			t.Fatalf("ed never showed a.txt: %q", head(&ed.app))
		}
	}
	g, top := gutter(len(ed.rows)), ed.tops()
	pos := at(corner, v.text.cell, g+5, top) // b.txt, from its rune 6
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse,
		Buttons: pointer.ButtonSecondary, Position: pos})
	router.Queue(pointer.Event{Kind: pointer.Release,
		Source: pointer.Mouse, Position: pos})
	for end := time.Now().Add(5 * time.Second); ; {
		settle()
		if strings.HasSuffix(head(&ed.app), "/b.txt") &&
			s.order[s.eye] == ed {
			break
		}
		if time.Now().After(end) {
			t.Fatalf("a right press on b.txt: ed at %q, "+
				"the eye on %s", head(&ed.app),
				s.order[s.eye].name)
		}
	}
}
