package main

import (
	"encoding/json"
	"io"
	"log"
	"time"
	"unicode"

	"sys/outline"
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

// row is a node as laid out; parent is a row's index, -1 at the top
type row struct {
	n             *outline.Node
	depth, parent int
}

type app struct {
	out    *json.Encoder
	tree   *outline.Node
	status string
	state  []string

	asking         bool
	prompt, answer string

	asTree     bool
	rows       []row
	head       string
	cur, top   int
	follow     bool // bring the cursor into sight on the next frame
	page, w, h int
	clicked    string
	clickedAt  time.Time
}

func (a *app) send(kind string, v any) { a.out.Encode(object{kind: v}) }

func (a *app) key(
	name, node string,
) {
	a.send("key", outline.Key{Name: name, Node: node})
}

func (a *app) isText() bool { return a.tree != nil && a.tree.Kind == "text" }

// zoom is the node the cursor is among the kids of, else nil
func zoom(n *outline.Node) *outline.Node {
	if n.In {
		return n
	}
	for _, k := range n.Kids {
		if z := zoom(k); z != nil {
			return z
		}
	}
	return nil
}

// layout makes the rows from the tree drawn last
func (a *app) layout() {
	a.rows, a.cur, a.follow = nil, 0, true
	if a.tree == nil {
		return
	}
	if a.isText() {
		a.head = a.tree.Text
		for i, k := range a.tree.Kids {
			a.rows = append(a.rows, row{k, 0, -1})
			for _, r := range k.Kids {
				if r.Cur {
					a.cur = i
				}
			}
		}
		return
	}
	shown := zoom(a.tree)
	if shown == nil || a.asTree {
		shown = a.tree
	}
	a.head = shown.ID
	here := 0
	var walk func(n *outline.Node, depth, parent int)
	walk = func(n *outline.Node, depth, parent int) {
		for _, k := range n.Kids {
			i := len(a.rows)
			a.rows = append(a.rows, row{k, depth, parent})
			if k.Cur {
				a.cur, here = i, -1
			} else if k.In && here >= 0 {
				here = i
			}
			if a.asTree && k.Open {
				walk(k, depth+1, i)
			}
		}
	}
	walk(shown, 0, -1)
	if here > 0 {
		a.cur = here
	}
}

// move and turn are the view's own: the program hears where the cursor landed
func (a *app) move(name string) bool {
	var d int
	switch name {
	case "j", "DOWN":
		d = 1
	case "k", "UP":
		d = -1
	case "PGDN":
		d = a.page
	case "PGUP":
		d = -a.page
	case "G":
		d = len(a.rows)
	default:
		return false
	}
	if len(a.rows) > 0 {
		a.key("", a.rows[max(0, min(a.cur+d, len(a.rows)-1))].n.ID)
	}
	return true
}

func folds(n *outline.Node, name string) bool {
	if n.Kind == "item" && len(n.Kids) == 0 {
		return false
	}
	return n.Kind == "dir" || n.Kind == "item" || name == "RIGHT"
}

func (a *app) turn(name string) bool {
	if !a.asTree || len(a.rows) == 0 {
		return false
	}
	r := a.rows[a.cur]
	switch name {
	case "ENTER", "RIGHT":
		if !folds(r.n, name) {
			return false
		}
		if r.n.Open {
			a.key("close", r.n.ID)
		} else {
			a.key("open", r.n.ID)
		}
	case "LEFT", "h":
		switch {
		case r.n.Open:
			a.key("close", r.n.ID)
		case r.parent >= 0:
			a.key("", a.rows[r.parent].n.ID)
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (a *app) pressed(name string) {
	switch {
	case name == "":
	case a.asking:
		switch name {
		case "ENTER", "ESC":
			a.asking = false
			text := a.answer
			if name == "ESC" {
				text = ""
			}
			a.send(
				"answer",
				object{"text": text, "ok": name == "ENTER"},
			)
		case "BACKSPACE":
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
	case name == "t":
		a.asTree = !a.asTree
		a.layout()
	case a.move(name), a.turn(name):
	case len(name) == 1 && unicode.IsDigit(rune(name[0])):
	default:
		a.key(name, "")
	}
}

func (a *app) click(id string) {
	if a.asking || a.isText() {
		return
	}
	name := ""
	if id == a.clicked && time.Since(a.clickedAt) < 400*time.Millisecond {
		name = "ENTER"
		for _, r := range a.rows {
			if r.n.ID == id && a.asTree && folds(r.n, name) {
				if r.n.Open {
					name = "close"
				} else {
					name = "open"
				}
			}
		}
	}
	a.clicked, a.clickedAt = id, time.Now()
	a.key(name, id)
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
