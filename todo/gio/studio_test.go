// the studio with no window: real fm -s and ed -s -f on a made folder, keys
// through gio's own input router, pictures from its offscreen gpu
package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestStudio(t *testing.T) {
	// SHOT names a folder to show for the picture; else a made one
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(
		filepath.Join(dir, "a.txt"),
		[]byte("alpha\n  one  a note\n  two\n"),
		0o644,
	)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("beta\n"), 0o644)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", programs(t)+":"+os.Getenv("PATH"))
	shot := os.Getenv("SHOT")
	if shot != "" {
		dir = shot
	}
	s, err := start(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, p := range s.panes {
			p.cmd.Process.Kill()
		}
	}()
	v, err := newView(s, true)
	if err != nil {
		t.Fatal(err)
	}
	const wide, high = 1500, 800
	var router input.Router
	var ops op.Ops
	done := false
	frame := func() {
		ops.Reset()
		gtx := layout.Context{Ops: &ops, Source: router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(wide, high))}
		_, over := v.draw(gtx)
		done = done || over
		router.Frame(&ops)
	}
	// settle draws until the programs have been quiet a while
	settle := func() {
		for quiet := 0; quiet < 15; quiet++ {
			if len(v.news) > 0 {
				quiet = 0
			}
			frame()
			time.Sleep(20 * time.Millisecond)
		}
	}
	gpu, err := headless.NewWindow(wide, high)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	shots := out(t)
	picture := func(name string) {
		frame()
		if err := gpu.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, wide, high))
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(shots, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		png.Encode(f, img)
	}
	settle()
	fm, ed := &s.panes[0].app, &s.panes[1].app
	char := func(r rune) {
		router.Queue(key.EditEvent{Text: string(r)})
		settle()
	}
	down := func(n key.Name, m key.Modifiers) {
		router.Queue(
			key.Event{Name: n, Modifiers: m, State: key.Press},
		)
		settle()
	}
	cur := func() string { return fm.rows[fm.cur].n.Text }
	has := func(what string, ok bool) {
		t.Helper()
		if !ok {
			t.Fatalf(
				"%s: fm at %q; ed %q status %q state %v",
				what,
				cur(),
				ed.head,
				ed.status,
				ed.state,
			)
		}
		t.Logf("ok %s", what)
	}
	if shot != "" {
		for i := 0; i < len(fm.rows) && cur() != "README.md"; i++ {
			char('j')
		}
		has(
			"ed follows fm to README.md",
			strings.HasSuffix(ed.head, "/README.md"),
		)
		picture("studio-sys.png")
		return
	}
	for i := 0; i < 5 && cur() != "a.txt"; i++ {
		char('j')
	}
	has("ed follows fm to a.txt", strings.HasSuffix(ed.head, "/a.txt"))
	char('j')
	has("ed follows fm to b.txt", strings.HasSuffix(ed.head, "/b.txt"))
	// a click on a.txt's row: the rows start five lines down
	at := f32.Pt(100, 6*cellH+5)
	router.Queue(
		pointer.Event{
			Kind:     pointer.Press,
			Source:   pointer.Mouse,
			Buttons:  pointer.ButtonPrimary,
			Position: at,
		},
	)
	router.Queue(
		pointer.Event{
			Kind:     pointer.Release,
			Source:   pointer.Mouse,
			Position: at,
		},
	)
	settle()
	has(
		"a click moves fm, and ed follows",
		cur() == "a.txt" && strings.HasSuffix(ed.head, "/a.txt"),
	)
	down("W", key.ModCtrl)
	has("ctrl-w moves the eye", s.eye == 1)
	char('i')
	char('Z')
	has(
		"typing goes to ed",
		ed.rows[0].n.Text == "Zalpha" && ed.state[0] == "insert",
	)
	picture("studio-insert.png")
	down(key.NameEscape, 0)
	down("W", key.ModCtrl)
	char('j')
	has(
		"unsaved ed holds as cur moves",
		strings.HasSuffix(ed.head, "/a.txt"),
	)
	picture("studio-held.png")
	char('q')
	has(
		"q in fm leaves ed, unsaved",
		s.panes[0].gone && !s.panes[1].gone && !done,
	)
	b, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
	has("the file is untouched", strings.HasPrefix(string(b), "alpha"))
}
