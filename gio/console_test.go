// a program's console in the frame: what fm runs comes to the front of its
// slot while it writes there, and fm's view comes back as fm draws again
package main

import (
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"sys/vt"
)

// TestConsole: f in the frame's fm runs its finder and fzy on fm's
// console, which comes to the front of fm's slot, the keys going to it:
// the line typed there is the pick, and fm, drawing again, lands on it,
// its view in front again; F's lint and pause come forward the same way,
// and a key there gives fm back; and o, which opens beside and writes
// nothing, leaves the console as it was. The finder, fzy and lint are
// stand-ins; fm, ed, studio, o and pause are sys's own, from its bin.
func TestConsole(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"alpha.txt", "beta.txt"} {
		os.WriteFile(filepath.Join(dir, f), []byte(f+"\n"), 0o644)
	}
	stand(t, map[string]string{
		"f": "echo beta.txt",
		"fzy": `list=$(cat); echo "fzy has $list" >/dev/tty; ` +
			`read pick </dev/tty; echo "$pick"`,
		"lint": `echo "lint $*"`,
	})
	s, router, until, _ := framed(t, dir)
	fm := s.panes["fm"]
	has := func(what string, ok bool) {
		t.Helper()
		if !ok {
			t.Fatalf("%s: fm at %q, %q; its console shows %q",
				what, cursor(&fm.app), fm.status,
				shows(fm.cons))
		}
		t.Logf("ok %s", what)
	}
	on := func(text string) bool {
		return strings.Contains(shows(fm.cons), text)
	}
	typed := func(text string) {
		router.Queue(key.EditEvent{Text: text})
		router.Queue(key.Event{Name: key.NameReturn, State: key.Press})
	}
	has("fm draws, its console behind",
		until(func() bool { return cursor(&fm.app) == "alpha.txt" }) &&
			fm.term == nil && fm.cons != nil)
	router.Queue(key.EditEvent{Text: "f"})
	has("f brings fm's console forward, fzy on it with the finder's line",
		until(func() bool {
			return fm.term == fm.cons && on("fzy has beta.txt")
		}))
	typed("beta.txt")
	has("the line typed there is the pick: fm lands on it, in front",
		until(func() bool {
			return fm.term == nil && cursor(&fm.app) == "beta.txt"
		}))
	router.Queue(key.EditEvent{Text: "F"})
	has("F's lint and pause come forward",
		until(func() bool {
			return fm.term == fm.cons && on("continue") &&
				on("lint -f "+filepath.Join(dir, "beta.txt"))
		}))
	typed("")
	has("a key there gives fm back",
		until(func() bool { return fm.term == nil }))
	// o ran once c's log names the row, by the verb o or e that ran
	// last, and its act is done once fm draws again
	was, log := shows(fm.cons), filepath.Join(os.Getenv("HOME"),
		".config/sys/log")
	router.Queue(key.EditEvent{Text: "o"})
	has("o runs, opening beside, and leaves the console as it was",
		until(func() bool {
			b, _ := os.ReadFile(log)
			return strings.Contains(string(b),
				filepath.Join(dir, "beta.txt")) && !fm.acting
		}) && fm.term == nil && shows(fm.cons) == was)
}

// TestBeside: what o opens beside, a viewer, lives on after the frame
// ends, though fm ran o on its console, which fm's end hangs up, and the
// window's closes: rc's &, which `sh/spawn` opens it with, gives it a
// process group of its own, and the hang-up reaches fm's alone. The
// viewer is a stand-in that says its pid.
func TestBeside(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "x.pdf"), nil, 0o644)
	said := filepath.Join(t.TempDir(), "pid")
	stand(t, map[string]string{"zathura": "echo $$ >" + said +
		"; exec sleep 30"})
	s, router, until, over := framed(t, dir)
	fm := s.panes["fm"]
	until(func() bool { return cursor(&fm.app) == "x.pdf" })
	router.Queue(key.EditEvent{Text: "o"})
	pid := 0
	if !until(func() bool {
		b, _ := os.ReadFile(said)
		pid, _ = strconv.Atoi(strings.TrimSpace(string(b)))
		return pid > 0
	}) {
		t.Fatal("o opened no viewer")
	}
	defer syscall.Kill(pid, syscall.SIGKILL)
	router.Queue(key.EditEvent{Text: "q"})
	if !until(over) {
		t.Fatal("q in fm did not end the frame")
	}
	for _, tabs := range s.tabs { // as the window goes, each pty closes
		for _, p := range tabs {
			if p.pty != nil {
				p.pty.Close()
			}
		}
	}
	time.Sleep(300 * time.Millisecond)
	if syscall.Kill(pid, 0) != nil {
		t.Error("the viewer o opened ended with the frame")
	}
}

// stand puts stand-ins first on the path, each a name and the shell it
// runs, sys's bin after them, in a home of the test's own
func stand(t *testing.T, ins map[string]string) {
	dir := t.TempDir()
	for name, body := range ins {
		os.WriteFile(filepath.Join(dir, name),
			[]byte("#!/bin/sh\n"+body+"\n"), 0o755)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", dir+":"+programs(t)+":"+os.Getenv("PATH"))
}

// framed lays studio's frame out on dir with no window, as TestStudio
// does: what is queued to router is pressed as a frame is drawn, until
// draws frames until ok holds or five seconds pass, and over is whether
// the frame ended
func framed(t *testing.T, dir string) (s *studio, router *input.Router,
	until func(ok func() bool) bool, over func() bool) {
	s, err := start(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, tabs := range s.tabs {
			for _, p := range tabs {
				if p.cmd != nil && p.cmd.Process != nil {
					p.cmd.Process.Kill()
				}
			}
		}
	})
	v, err := newWindow(s, true)
	if err != nil {
		t.Fatal(err)
	}
	router = new(input.Router)
	var ops op.Ops
	done := false
	frame := func() {
		ops.Reset()
		_, end := v.draw(layout.Context{Ops: &ops,
			Source:      router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Exact(image.Pt(1500, 800))})
		done = done || end
		router.Frame(&ops)
	}
	until = func(ok func() bool) bool {
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
	return s, router, until, func() bool { return done }
}

// shows is the text on a terminal's screen, its rows a line each
func shows(t *vt.Term) string {
	if t == nil {
		return ""
	}
	_, rows := t.Size()
	var b strings.Builder
	for y := range rows {
		for _, c := range t.Line(y) {
			b.WriteRune(glyph(c))
		}
		b.WriteString("\n")
	}
	return b.String()
}
