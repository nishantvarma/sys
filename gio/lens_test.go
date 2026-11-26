// ed's lenses in the window: a real ed -s at the other end of the pipe,
// keys and presses through gio's own router, as on the terminal
package main

import (
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// stripText is a strip as text, its runs joined
func stripText(k *strip) string {
	var b strings.Builder
	for _, r := range k.runs {
		b.WriteString(r.text)
	}
	return b.String()
}

// TestLens: ed in the window shows its lenses as the terminal does: the
// bar and its rule over the text; Tab turns to the outline, sys/view's
// rows under the same bar; a press on a lens's name on the bar shows that
// lens, and esc is source again.
func TestLens(t *testing.T) {
	note := filepath.Join(t.TempDir(), "note")
	os.WriteFile(note, []byte("t\nup\n  a\n    → b  go\nlow\n  b\n"),
		0o644)
	bin := programs(t)
	c := exec.Command(bin+"/ed", "-s", note)
	c.Env = []string{"HOME=" + t.TempDir(),
		"PATH=" + bin + ":" + os.Getenv("PATH")}
	to, _ := c.StdinPipe()
	from, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	s := piped(from, to)
	v, err := newWindow(s, false)
	if err != nil {
		t.Fatal(err)
	}
	var router input.Router
	var ops op.Ops
	one := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	settle := func() {
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
	p := s.order[0]
	lines := func() []string {
		ls := p.list()
		if p.isText() {
			ls = p.source()
		}
		var out []string
		for _, l := range ls {
			out = append(out, stripText(l))
		}
		return out
	}
	barred := func(what string) []string {
		t.Helper()
		ls := lines()
		if len(ls) < 2 || !strings.HasPrefix(ls[0],
			"  source  outline  render") ||
			strings.Trim(ls[1], "─") != "" {
			t.Fatalf("%s: no bar over a rule:\n%s", what,
				strings.Join(ls, "\n"))
		}
		return ls
	}
	settle()
	if ls := barred("source"); !p.isText() ||
		strings.TrimRight(ls[2], " ") != "   1  t" {
		t.Fatalf("source is not ed's text under the bar:\n%s",
			strings.Join(ls, "\n"))
	}
	router.Queue(key.Event{Name: key.NameTab, State: key.Press})
	settle()
	ls := barred("outline")
	if p.tree.Kind != "outline" ||
		!strings.Contains(strings.Join(ls, "\n"), "    → b  go") {
		t.Fatalf("Tab drew %q:\n%s", p.tree.Kind,
			strings.Join(ls, "\n"))
	}
	x := strings.Index(ls[0], "render") * v.text.cell
	v.pressedAt(image.Pt(x+1, 1), v.text.cell, cellH)
	settle()
	if p.tree.Kind != "picture" {
		t.Fatalf("a press on render drew %q", p.tree.Kind)
	}
	barred("render")
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	settle()
	if !p.isText() {
		t.Fatalf("esc in the render drew %q, not source", p.tree.Kind)
	}
}
