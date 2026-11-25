// the window mode: a real fm -s at the other end of the pipe, keys through
// gio's input router
package main

import (
	"image"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestPiped(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir+"/alpha/inner", 0o755)
	os.Mkdir(dir+"/beta", 0o755)
	bin := programs(t)
	c := exec.Command(bin+"/fm", "-s", dir)
	c.Env = []string{
		"HOME=" + t.TempDir(),
		"PATH=" + bin + ":" + os.Getenv("PATH"),
	}
	toFm, _ := c.StdinPipe()
	fromFm, _ := c.StdoutPipe()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	var out io.Writer = toFm
	s := piped(fromFm, out)
	v, err := newView(s, false)
	if err != nil {
		t.Fatal(err)
	}
	var router input.Router
	var ops op.Ops
	settle := func() {
		for quiet := 0; quiet < 10; quiet++ {
			if len(v.news) > 0 {
				quiet = 0
			}
			ops.Reset()
			v.draw(
				layout.Context{
					Ops:    &ops,
					Source: router.Source(),
					Metric: unit.Metric{
						PxPerDp: 1,
						PxPerSp: 1,
					},
					Constraints: layout.Exact(
						image.Pt(760, 560),
					),
				},
			)
			router.Frame(&ops)
			time.Sleep(20 * time.Millisecond)
		}
	}
	p := s.panes[0]
	at := func(what, head, cur string) {
		t.Helper()
		got := ""
		if len(p.rows) > 0 {
			got = p.rows[p.cur].n.Text
		}
		if p.head != head || got != cur {
			t.Fatalf(
				"%s: at %q %q, want %q %q",
				what,
				p.head,
				got,
				head,
				cur,
			)
		}
		t.Logf("ok %s", what)
	}
	press := func(n key.Name) {
		router.Queue(key.Event{Name: n, State: key.Press})
		settle()
	}
	typed := func(s string) {
		router.Queue(key.EditEvent{Text: s})
		settle()
	}
	settle()
	at("start", dir, "alpha")
	typed("j")
	at("j", dir, "beta")
	press(key.NameUpArrow)
	at("up", dir, "alpha")
	press(key.NameReturn)
	at("enter", dir+"/alpha", "inner")
	typed("/")
	if !p.asking {
		t.Fatalf("/: no prompt, status %q", p.status)
	}
	press(key.NameEscape)
	at("prompt opened and cancelled", dir+"/alpha", "inner")
}
