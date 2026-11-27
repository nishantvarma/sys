package main

import (
	"image"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// TestTerminal: the frame's panel runs a shell in a terminal of the
// window's own: what is typed reaches it and what it writes shows, as
// the window draws it, and a press on its last row, which the window's
// height leaves part of, is on its screen. + opens a second, l new typed
// in a terminal a third, and l outline the lens view's tab by that name;
// a number is a terminal's place, after a close and a new too; a
// terminal's exit gives its place to the next, though a job it left
// holds its pty; × closes the one shown, the panel's bar naming them all
// the while; and a shell that writes without end keeps the window
// drawing, and ctrl-c reaching it.
func TestTerminal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/sh")
	t.Setenv("PATH", programs(t)+":"+os.Getenv("PATH"))
	s, err := start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, tabs := range s.tabs {
			for _, p := range tabs {
				if p.cmd != nil {
					p.cmd.Process.Kill()
				}
			}
		}
	}()
	v, err := newWindow(s, true)
	if err != nil {
		t.Fatal(err)
	}
	var router input.Router
	var ops op.Ops
	const high = 805 // five pixels past the last whole row
	frame := func() {
		ops.Reset()
		gtx := layout.Context{Ops: &ops, Source: router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1500, high))}
		v.draw(gtx)
		router.Frame(&ops)
	}
	// until draws frames until ok holds, five seconds at most
	until := func(ok func() bool) bool {
		for end := time.Now().Add(5 * time.Second); ; {
			frame()
			if ok() {
				return true
			}
			if time.Now().After(end) {
				return false
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	panel := func() *pane { return s.panes[s.panel] }
	// screen is the panel's terminal as the window draws it, a strip a
	// row (row)
	screen := func() string {
		var lines []string
		_, rows := panel().term.Size()
		for y := range rows {
			var b strings.Builder
			k := row(panel().term.Line(y), false)
			for _, r := range k.runs {
				b.WriteString(r.text)
			}
			line := strings.TrimRight(b.String(), " ")
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}
	last := func() string {
		lines := strings.Split(strings.TrimRight(screen(), "\n"), "\n")
		return lines[len(lines)-1]
	}
	prompt := func() bool {
		return !panel().gone && strings.Contains(screen(), "$")
	}
	typing := func(text string) {
		router.Queue(key.EditEvent{Text: text},
			key.Event{Name: key.NameReturn, State: key.Press})
	}
	// bar is the panel's bar as drawn: its words, and where to press
	// each
	bar := func() ([]string, map[string]f32.Point) {
		var words []string
		at := map[string]f32.Point{}
		for _, b := range v.bars {
			if b.slot != s.panel {
				continue
			}
			for _, c := range b.chips {
				words = append(words, c.Name)
				x := b.area.Min.X + (c.From+1)*v.text.cell
				at[c.Name] = f32.Pt(float32(x),
					float32(b.area.Min.Y+cellH/2))
			}
		}
		return words, at
	}
	press := func(word string) {
		_, at := bar()
		router.Queue(pointer.Event{Kind: pointer.Press,
			Source: pointer.Mouse, Buttons: pointer.ButtonPrimary,
			Position: at[word]},
			pointer.Event{Kind: pointer.Release,
				Source: pointer.Mouse, Position: at[word]})
	}
	tabs := func() []string {
		var names []string
		for _, p := range s.tabs[s.panel] {
			names = append(names, p.tab)
		}
		return names
	}
	shows := func(want []string, tab string) bool {
		return until(func() bool {
			return slices.Equal(tabs(), want) &&
				panel().tab == tab && prompt()
		})
	}

	if !until(prompt) {
		t.Fatalf("the panel's shell gave no prompt: %q", screen())
	}
	s.focus(slices.Index(s.order, panel()))
	typing("echo hi")
	if !until(func() bool { return strings.Contains(screen(), "\nhi") }) {
		t.Errorf("echo hi typed, and the terminal shows %q", screen())
	}
	if words, _ := bar(); !slices.Equal(words, []string{"1", "new"}) {
		t.Errorf("the panel's bar says %q", words)
	}
	bottom := f32.Pt(100, high-2)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse,
		Buttons: pointer.ButtonPrimary, Position: bottom},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse,
			Position: bottom})
	frame() // no index past the screen's last row
	press("new")
	if !shows([]string{"shell", "2"}, "2") {
		t.Fatalf("+ pressed, and the panel holds %q, shows %q", tabs(),
			panel().tab)
	}
	if words, _ := bar(); !slices.Equal(words,
		[]string{"1", "2", "new", "close"}) {
		t.Errorf("with two, the panel's bar says %q", words)
	}
	typing("l new")
	if !shows([]string{"shell", "2", "3"}, "3") {
		t.Fatalf("l new typed, and the panel holds %q, shows %q",
			tabs(), panel().tab)
	}
	typing("l outline")
	if !until(func() bool { return s.panes["view"].tab == "outline" }) {
		t.Errorf("l outline typed in a terminal, and the lens view"+
			" shows %q", s.panes["view"].tab)
	}
	s.focusOn(panel())
	press("2")
	press("close")
	press("new")
	if !shows([]string{"shell", "3", "4"}, "4") {
		t.Fatalf("2 closed and one new, and the panel holds %q,"+
			" shows %q", tabs(), panel().tab)
	}
	press("2")
	if !shows([]string{"shell", "3", "4"}, "3") {
		t.Errorf("a press on 2, and the panel shows %q, not the"+
			" second", panel().tab)
	}
	typing("sleep 30 & exit")
	if !shows([]string{"shell", "4"}, "4") {
		t.Fatalf("a terminal exited, a job left, and the panel holds"+
			" %q, shows %q", tabs(), panel().tab)
	}
	press("close")
	if !shows([]string{"shell"}, "shell") {
		t.Errorf("× pressed, and the panel holds %q, shows %q",
			tabs(), panel().tab)
	}
	typing("yes")
	until(func() bool { return strings.Contains(screen(), "y\ny") })
	router.Queue(key.Event{Name: "C", Modifiers: key.ModCtrl,
		State: key.Press})
	if !until(func() bool { return strings.HasSuffix(last(), "$") }) {
		t.Errorf("yes, and ctrl-c never stopped it: %q", last())
	}
}
