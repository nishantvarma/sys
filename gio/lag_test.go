// where a key's time goes, on a real tree: LAG=dir go test -run Lag -v
package main

import (
	"image"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestLag(t *testing.T) {
	dir := os.Getenv("LAG")
	if dir == "" {
		t.Skip("set LAG to a folder")
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
	gpu, err := headless.NewWindow(1360, 800)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	var router input.Router
	var ops op.Ops
	var build, render []time.Duration
	frame := func(timed bool) {
		a := time.Now()
		ops.Reset()
		v.draw(layout.Context{Ops: &ops, Source: router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1360, 800))})
		router.Frame(&ops)
		b := time.Now()
		gpu.Frame(&ops)
		if timed {
			build, render = append(
				build,
				b.Sub(a),
			), append(
				render,
				time.Since(b),
			)
		}
	}
	// take handles news as it comes, until none for quiet
	take := func(quiet time.Duration, each func(news)) {
		for {
			select {
			case e := <-s.events:
				s.handle(e)
				each(e)
			case <-time.After(quiet):
				return
			}
		}
	}
	frame(false)
	take(500*time.Millisecond, func(news) {})
	frame(false)
	fm, ed := s.panes["fm"], s.panes["view"]
	var fmT, edT []time.Duration
	for i := 0; fm.tree != nil && i < len(fm.tree.Kids)-1; i++ {
		was := head(&ed.app)
		t0 := time.Now()
		var gotFm, gotEd time.Duration
		fm.pressed("j")
		take(300*time.Millisecond, func(e news) {
			if e.m != nil && e.m.Draw != nil {
				if e.p == fm && gotFm == 0 {
					gotFm = time.Since(t0)
				}
				moved := head(&ed.app) != was
				if e.p == ed && gotEd == 0 && moved {
					gotEd = time.Since(t0)
				}
			}
		})
		fmT = append(fmT, gotFm)
		if gotEd > 0 && strings.HasPrefix(head(&ed.app), dir) {
			edT = append(edT, gotEd)
		}
		frame(true)
	}
	say := func(what string, d []time.Duration) {
		slices.Sort(d)
		if len(d) > 0 {
			t.Logf(
				"%-34s median %8v  worst %8v  (%d)",
				what,
				d[len(d)/2].Round(100*time.Microsecond),
				d[len(d)-1].Round(100*time.Microsecond),
				len(d),
			)
		}
	}
	say("j to fm's redraw arriving", fmT)
	say("j to ed showing the new file", edT)
	say("gio: build a frame (cpu)", build)
	say("gio: render it, offscreen", render)
}
