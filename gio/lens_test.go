// the lens in the window: a real studio show -s at the other end of the
// pipe, keys and presses through gio's own router, as on the terminal
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

// TestLens: the lens in the window shows as on the terminal: the note cur
// names, as its outline's rows, the lens shown said in its state; Tab
// turns to the render, the note's picture, and again to the diff.
func TestLens(t *testing.T) {
	note := filepath.Join(t.TempDir(), "note")
	os.WriteFile(note, []byte("t\nup\n  a\n    → b  go\nlow\n  b\n"),
		0o644)
	home := t.TempDir()
	cur := filepath.Join(home, ".config", "sys", "cur", "none")
	os.MkdirAll(filepath.Dir(cur), 0o755)
	os.WriteFile(cur, []byte(note+":1\n"), 0o644)
	bin := programs(t)
	c := exec.Command(bin+"/studio", "show", "-s")
	c.Env = []string{"HOME=" + home,
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
	// shows is the lens shown, as its state says it, and what it drew
	shows := func(what string) string {
		t.Helper()
		var ls []string
		for _, l := range p.list() {
			ls = append(ls, stripText(l))
		}
		if len(p.state) == 0 || p.state[0] != what {
			t.Fatalf("the lens says %v, not %s", p.state, what)
		}
		return strings.Join(ls, "\n")
	}
	settle()
	if drew := shows("outline"); p.tree.Kind != "outline" ||
		!strings.Contains(drew, "    → b  go") {
		t.Fatalf("the outline drew %q:\n%s", p.tree.Kind, drew)
	}
	router.Queue(key.Event{Name: key.NameTab, State: key.Press})
	settle()
	if shows("render"); p.tree.Kind != "picture" {
		t.Fatalf("Tab drew %q, not the render", p.tree.Kind)
	}
	router.Queue(key.Event{Name: key.NameTab, State: key.Press})
	settle()
	shows("diff")
}
