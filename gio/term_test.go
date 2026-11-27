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
// window's own: what is typed reaches it and what it writes shows. +
// opens a second, l new typed in a terminal a third, a terminal's exit
// gives its place to the next, and × closes the one shown, the panel's
// bar naming them all the while.
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
	frame := func() {
		ops.Reset()
		gtx := layout.Context{Ops: &ops, Source: router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1500, 800))}
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
	screen := func() string {
		var lines []string
		_, rows := panel().term.Size()
		for y := range rows {
			var b strings.Builder
			for _, c := range panel().term.Line(y) {
				b.WriteRune(glyph(c))
			}
			line := strings.TrimRight(b.String(), " ")
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
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
				words = append(words, c.Lens)
				x := b.area.Min.X + (c.From+1)*v.text.cell
				at[c.Lens] = f32.Pt(float32(x),
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
	typing("exit")
	if !shows([]string{"shell", "2"}, "shell") {
		t.Fatalf("a terminal exited, and the panel holds %q, shows %q",
			tabs(), panel().tab)
	}
	press("close")
	if !shows([]string{"2"}, "2") {
		t.Errorf("× pressed, and the panel holds %q, shows %q",
			tabs(), panel().tab)
	}
}
