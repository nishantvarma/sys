package main

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"sys/view"
)

// out is where a test's pictures go: $OUT to keep them
func out(t *testing.T) string {
	if d := os.Getenv("OUT"); d != "" {
		return d
	}
	return t.TempDir()
}

// programs is sys's bin, where ./build puts fm and ed
func programs(t *testing.T) string {
	bin, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	return bin
}

// said is a line of sys/view's as text, its runs joined
func said(rs []view.Run) string {
	var b strings.Builder
	for _, r := range rs {
		b.WriteString(r.Text)
	}
	return b.String()
}

// head is the path of what a program drew last: a text's file, or the
// folder fm stands in
func head(a *app) string {
	switch {
	case a.tree == nil:
		return ""
	case a.isText():
		return a.tree.Text
	}
	return a.tree.ID
}

// cursor is the text of the row the cursor is on, in what a program drew
// last: fm's row, or ed's line
func cursor(a *app) string {
	if a.tree == nil {
		return ""
	}
	for _, k := range a.tree.Kids {
		if k.Cur {
			return k.Text
		}
	}
	if a.isText() && a.cur < len(a.rows) {
		return a.rows[a.cur].Text
	}
	return ""
}

func gtxFor(ops *op.Ops, w, h int) layout.Context {
	return layout.Context{
		Ops:         ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Exact(image.Pt(w, h)),
	}
}

// TestHelp: ? in the window shows ed's keys, as nodes sys/view lays out
// as rows, and esc closes them, back to the text: help needs no
// terminal (`design/view:/help is nodes/`).
func TestHelp(t *testing.T) {
	note := filepath.Join(t.TempDir(), "note")
	os.WriteFile(note, []byte("alpha\n"), 0o644)
	bin := programs(t)
	env := []string{"HOME=" + t.TempDir(),
		"PATH=" + bin + ":" + os.Getenv("PATH")}
	v, router, settle := alone(t, env, bin+"/ed", "-s", note)
	settle()
	p := v.s.order[0]
	router.Queue(key.EditEvent{Text: "?"})
	settle()
	rows := ""
	for _, l := range p.list() {
		rows += stripText(l) + "\n"
	}
	if p.tree == nil || p.tree.Kind != "help" ||
		!strings.Contains(rows, "line start") {
		t.Fatalf("? showed:\n%s", rows)
	}
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	settle()
	if !p.isText() {
		t.Errorf("esc left %q, not ed's text", p.tree.Kind)
	}
}
