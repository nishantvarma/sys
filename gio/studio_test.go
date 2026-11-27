// the studio with no window: real fm -s and ed -s -f on a made folder, keys
// through gio's own input router, pictures from its offscreen gpu
package main

import (
	"image"
	"image/png"
	"os"
	"os/exec"
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
			if p.cmd != nil {
				p.cmd.Process.Kill()
			}
		}
	}()
	v, err := newWindow(s, true)
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
	fm, ed := &s.panes["fm"].app, &s.panes["view"].app
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
	cur := func() string { return cursor(fm) }
	has := func(what string, ok bool) {
		t.Helper()
		if !ok {
			t.Fatalf(
				"%s: fm at %q; ed %q status %q state %v",
				what,
				cur(),
				head(ed),
				ed.status,
				ed.state,
			)
		}
		t.Logf("ok %s", what)
	}
	if shot != "" {
		for i := 0; i < 99 && cur() != "README.md"; i++ {
			char('j')
		}
		has(
			"ed follows fm to README.md",
			strings.HasSuffix(head(ed), "/README.md"),
		)
		picture("studio-sys.png")
		return
	}
	for i := 0; i < 5 && cur() != "a.txt"; i++ {
		char('j')
	}
	has("ed follows fm to a.txt", strings.HasSuffix(head(ed), "/a.txt"))
	char('j')
	has("ed follows fm to b.txt", strings.HasSuffix(head(ed), "/b.txt"))
	char('t')
	// a card names sub/ and says it is a folder on the line under
	cards, thumbs := fm.v.Lines(), false
	for i := 0; i+1 < len(cards); i++ {
		thumbs = thumbs || strings.Contains(said(cards[i]), "sub/") &&
			strings.Contains(said(cards[i+1]), "folder")
	}
	has("t lays fm out as thumbs, by sys/view", thumbs)
	char('t')
	// a click on a.txt's row: the rows start five lines down, . and ..
	// first, then sub
	at := f32.Pt(100, 8*cellH+5)
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
		cur() == "a.txt" && strings.HasSuffix(head(ed), "/a.txt"),
	)
	down("W", key.ModCtrl)
	has("ctrl-w moves the eye", s.order[s.eye].name == "view")
	char('i')
	char('Z')
	has(
		"typing goes to ed",
		ed.rows[0].Text == "Zalpha" && ed.state[0] == "insert",
	)
	picture("studio-insert.png")
	// ed's status line, the bottom of its slot: its first word the mode
	press := func(at f32.Point) {
		router.Queue(pointer.Event{Kind: pointer.Press,
			Source: pointer.Mouse, Buttons: pointer.ButtonPrimary,
			Position: at})
		router.Queue(pointer.Event{Kind: pointer.Release,
			Source: pointer.Mouse, Position: at})
		settle()
	}
	// ed's corner, as the window drew it: under the lens view's bar
	var corner image.Point
	for _, h := range v.hits {
		if h.p == s.panes["view"] {
			corner = h.area.Min
		}
	}
	word := f32.Pt(float32(corner.X+12),
		float32(corner.Y+(s.panes["view"].high-1)*cellH+5))
	press(word)
	has("a click on the mode leaves insert", ed.state[0] == "normal")
	press(word)
	has("and goes back in", ed.state[0] == "insert")
	down(key.NameEscape, 0)
	// the lens view's tabs: its bar names them; ed's Tab runs l, which
	// asks the window by its ask file, and the outline shows; a press
	// on a name on the bar shows that one; alt-t hides the panel and
	// shows it again, the eye with it, and so does studio panel, run
	// with no tmux, asking the window; and the eye follows the pointer
	bar, _ := s.tabBar("view", 80)
	has("the lens view's bar names its tabs", bar != nil &&
		strings.Contains(stripText(bar),
			"source  outline  render  diff"))
	// until settles the window until ok holds, or five seconds pass
	until := func(ok func() bool) bool {
		for end := time.Now().Add(5 * time.Second); ; {
			settle()
			if ok() {
				return true
			}
			if time.Now().After(end) {
				return false
			}
		}
	}
	shows := func(tab string) bool {
		return until(func() bool { return s.panes["view"].tab == tab })
	}
	down(key.NameTab, 0)
	has("Tab in ed shows the outline, the lens view's next tab",
		shows("outline"))
	// the first chip naming render is its name; the last, tab, names
	// the next, which render is now too
	var on image.Point
	for _, b := range v.bars {
		for _, c := range b.chips {
			first := on == (image.Point{})
			if b.slot == "view" && c.Name == "render" && first {
				x := b.area.Min.X + c.From*v.text.cell
				on = image.Pt(x+1, b.area.Min.Y+5)
			}
		}
	}
	press(f32.Pt(float32(on.X), float32(on.Y)))
	has("a press on render, on the bar, shows it", shows("render"))
	eye := func() *pane { return s.order[s.eye] }
	down("T", key.ModAlt)
	has("alt-t hides the panel", s.hidden)
	down("T", key.ModAlt)
	has("and shows it again, the eye on its terminal", !s.hidden &&
		eye().name == s.panel && eye().term != nil)
	panel := func(args ...string) {
		c := exec.Command("studio", append([]string{"panel"},
			args...)...)
		c.Env = s.env
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("studio panel: %v: %s", err, out)
		}
	}
	panel("hide")
	has("studio panel hide hides it, the eye moving off it",
		until(func() bool { return s.hidden }) &&
			eye().name != s.panel)
	panel()
	has("and bare, shows it again, the eye on it",
		until(func() bool { return !s.hidden }) &&
			eye().name == s.panel)
	down("W", key.ModCtrl)
	has("ctrl-w goes on, round to fm", eye().name == "fm")
	move := func(at f32.Point) {
		router.Queue(pointer.Event{Kind: pointer.Move,
			Source: pointer.Mouse, Position: at})
		settle()
	}
	move(word)
	has("the pointer on ed gives it the eye", eye().name == "view")
	move(at)
	has("and on fm, fm", eye().name == "fm")
	char('j')
	has(
		"unsaved ed holds as cur moves",
		strings.HasSuffix(head(ed), "/a.txt"),
	)
	picture("studio-held.png")
	char('q')
	has(
		"q in fm, ed unsaved, ends nothing and says so",
		!s.panes["fm"].gone && !s.panes["view"].gone && !done &&
			strings.HasPrefix(s.panes["fm"].status,
				"unsaved in ed: a.txt"),
	)
	char('q')
	has("a second q ends the frame", s.panes["fm"].gone && done)
	b, _ := os.ReadFile(filepath.Join(dir, "a.txt"))
	has("the file is untouched", strings.HasPrefix(string(b), "alpha"))
}
