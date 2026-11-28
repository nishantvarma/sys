package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"sys/outline"
	"sys/tty"
	"sys/view"
)

type object = map[string]any

type in struct {
	Draw *struct {
		Tree   *outline.Node
		Status string
		State  []string
	}
	Ask  *struct{ Prompt, Def string }
	Copy *struct{ Text string }
}

// app is one program as the window holds it: what it drew last, a
// prompt it asked, and how it is laid out. What is no text, fm's folder,
// sys/view lays out, as on the terminal, moving the cursor
// itself and handing the program the rest; a text, ed's, is drawn here,
// a line a row (source).
type app struct {
	out    *json.Encoder
	tree   *outline.Node
	status string
	state  []string
	// acting: sent a key it has not drawn since, so that what writes on
	// its console meanwhile is what it runs (handle)
	acting bool

	asking         bool
	prompt, answer string

	v         *view.Term
	rows      []*outline.Node // a text's lines
	cur       int             // the line a text's cursor is in
	clicked   string
	clickedAt time.Time
}

// newApp is a program as the window holds it, what is no text laid out by
// sys/view, which says how each kind looks, as on the terminal.
func newApp(out *json.Encoder) app {
	home, _ := os.UserHomeDir()
	return app{out: out, v: view.New(nil, home)}
}

func (a *app) send(kind string, v any) {
	a.acting = a.acting || kind == "key"
	a.out.Encode(object{kind: v})
}

func (a *app) key(name, node string) {
	a.send("key", outline.Key{Name: name, Node: node})
}

func (a *app) isText() bool { return a.tree != nil && a.tree.Kind == "text" }

// layout takes the tree drawn last: a text's lines, the cursor's found
// among them, or the view's to lay out
func (a *app) layout() {
	a.rows, a.cur = nil, 0
	if a.tree == nil {
		return
	}
	if !a.isText() {
		a.v.Draw(a.tree, a.status, a.state)
		return
	}
	a.rows = a.tree.Kids
	for i, k := range a.rows {
		for _, r := range k.Runs { // a line holds the cursor in a run
			if r.Cur {
				a.cur = i
			}
		}
	}
}

// pressed is a key named as tty names it: a prompt's while one is asked,
// a text's program's, else the view's first
func (a *app) pressed(name string) {
	switch {
	case name == "":
	case a.asking:
		switch name {
		case tty.KEnter, tty.KEsc:
			a.asking = false
			text := a.answer
			if name == tty.KEsc {
				text = ""
			}
			a.send(
				"answer",
				object{"text": text, "ok": name == tty.KEnter},
			)
		case tty.KBack:
			if r := []rune(a.answer); len(r) > 0 {
				a.answer = string(r[:len(r)-1])
			}
		default:
			if len([]rune(name)) == 1 {
				a.answer += name
			}
		}
	case a.isText():
		a.key(name, "")
	default:
		k := tty.Key{Name: name}
		if r := []rune(name); len(r) == 1 {
			k = tty.Key{Rune: r[0]}
		}
		if ev, theirs := a.v.Handle(k); theirs {
			a.send("key", ev)
		}
	}
}

// click is a press on node id, a text's line at its rune at: the cursor
// goes there, and a second press on a row soon after is enter there
func (a *app) click(id string, at int) {
	if a.asking || id == "" {
		return
	}
	if a.isText() {
		a.send("key", outline.Key{Node: id, At: at})
		return
	}
	name := ""
	if id == a.clicked && time.Since(a.clickedAt) < 400*time.Millisecond {
		name = tty.KEnter
	}
	a.clicked, a.clickedAt = id, time.Now()
	a.key(name, id)
}

// modeWord is the key a press on a text's status line is, at column x
// from 0: on the mode, its first word, esc out of insert and i into it,
// as on the terminal (`ed/4-view/30-modeword.go`); "" elsewhere
func (a *app) modeWord(x int) string {
	if len(a.state) == 0 || x > len(a.state[0]) {
		return ""
	}
	if a.state[0] == "insert" {
		return tty.KEsc
	}
	return "i"
}

// feed hands each message to f until the stream ends
func feed(r io.Reader, f func(in)) {
	d := json.NewDecoder(r)
	for {
		var m in
		if err := d.Decode(&m); err != nil {
			if err != io.EOF {
				log.Print(err)
			}
			return
		}
		f(m)
	}
}
