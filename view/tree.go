package view

import (
	"fmt"
	"slices"
	"strings"

	"sys/outline"
	"sys/tty"
)

// An outline is a tree the view lays out as rows that fold, as the page
// folds its nodes: a part's rows are the parts under it, but its lead,
// which says what it is, beside it on the cursor's row. The view unfolds
// the way to the program's cursor (reveal), and → and ← unfold and fold
// the rest (fold). Each part is drawn by its kind (part), and a kind the
// view does not know as its text (`design/view:/Kinds, never looks/`).

// isOutline is whether the tree drawn last is an outline.
func (v *Term) isOutline() bool {
	return v.last != nil && v.last.Kind == "outline"
}

// lead is what is said of a part, which stands beside it: its first
// part, when that is a paragraph; nil for none.
func lead(n *outline.Node) *outline.Node {
	if len(n.Kids) > 0 && n.Kids[0].Kind == "para" {
		return n.Kids[0]
	}
	return nil
}

// declares is whether a part declares a name, as a type or a func does:
// a run of its text is the name it declares.
func declares(n *outline.Node) bool {
	return slices.ContainsFunc(n.Runs,
		func(r *outline.Node) bool { return r.Kind == "name" })
}

// held is the rows a part holds: the parts under it, but its lead.
func held(n *outline.Node) []*outline.Node {
	if lead(n) != nil {
		return n.Kids[1:]
	}
	return n.Kids
}

// sign is what stands before an item's or a step's text, and its colour:
// a dot, or a box, ticked when the step is done; "" for any other kind.
func sign(n *outline.Node) (string, string) {
	switch {
	case n.Kind == "item":
		return "• ", tty.Faint
	case n.Kind != "step":
		return "", ""
	case slices.Contains(n.Tags, "done"):
		return "☑ ", tty.Plus
	}
	return "☐ ", tty.Faint
}

// column is how wide the names of the rows among ns stand, the widest's,
// so that what each says lines up beside it.
func column(ns []*outline.Node) int {
	w := 0
	for _, n := range ns {
		if n.Kind == "row" {
			w = max(w, tty.Width(name(n)))
		}
	}
	return w
}

// apart is whether a blank line stands between two parts side by side in
// an outline, prev and n, as a doc sets them apart (gap): parts of two
// kinds, a list's items and steps one, and a part that holds rows on
// both sides, while leaves of a kind stand together, as a table does.
func apart(prev, n *outline.Node) bool {
	list := func(k string) bool { return k == "item" || k == "step" }
	same := prev.Kind == n.Kind || list(prev.Kind) && list(n.Kind)
	return !same || len(held(prev)) > 0 || len(held(n)) > 0
}

// index notes the part each part is under, in the tree drawn last.
func (v *Term) index() {
	v.up = map[*outline.Node]*outline.Node{}
	var walk func(p *outline.Node)
	walk = func(p *outline.Node) {
		for _, k := range p.Kids {
			v.up[k] = p
			walk(k)
		}
	}
	walk(v.last)
}

// marked is the part the program drew the cursor on, nil for none.
func (v *Term) marked() *outline.Node {
	var find func(ns []*outline.Node) *outline.Node
	find = func(ns []*outline.Node) *outline.Node {
		for _, n := range ns {
			if n.Cur {
				return n
			}
			if c := find(n.Kids); c != nil {
				return c
			}
		}
		return nil
	}
	return find(v.last.Kids)
}

// rowOf is the row a part is drawn on: its own, or, for a lead, which
// stands on its part's row, that part's.
func (v *Term) rowOf(n *outline.Node) *outline.Node {
	for p := v.up[n]; p != nil && p != v.last && n == lead(p); {
		n, p = p, v.up[p]
	}
	return n
}

// reveal unfolds the way to part n, each part it is under.
func (v *Term) reveal(n *outline.Node) {
	for p := v.up[v.rowOf(n)]; p != nil && p != v.last; p = v.up[p] {
		v.unfolded[p.ID] = true
	}
}

// depth is how many parts n is under.
func (v *Term) depth(n *outline.Node) int {
	d := 0
	for p := v.up[n]; p != nil && p != v.last; p = v.up[p] {
		d++
	}
	return d
}

// cursor is the row the cursor is on: the one a move went to, till the
// program draws again, else the row of the part the program drew it on;
// nil for none.
func (v *Term) cursor() *outline.Node {
	if v.cur != "" {
		for _, r := range v.rows() {
			if r.ID == v.cur {
				return r
			}
		}
	}
	if c := v.marked(); c != nil {
		return v.rowOf(c)
	}
	return nil
}

// fold does → or ←, as the page does: → unfolds the row at the cursor
// when it holds rows, and on one unfolded goes in, to its first; ← folds
// one unfolded, and on one folded, or holding none, goes to the row it is
// under and folds that, or, on a top row, to the row before. It is the id
// of the row the cursor goes to; moved is false when it stays.
func (v *Term) fold(name string) (to string, moved bool) {
	rows := v.rows()
	if len(rows) == 0 {
		return "", false
	}
	i := v.curRow(rows)
	n, in := rows[i], held(rows[i])
	unfolded := v.unfolded[n.ID] && len(in) > 0
	switch {
	case name == tty.KRight && len(in) == 0:
	case name == tty.KRight && !unfolded:
		v.unfolded[n.ID] = true
	case name == tty.KRight:
		return in[0].ID, true
	case unfolded:
		delete(v.unfolded, n.ID)
	case v.up[n] != v.last:
		delete(v.unfolded, v.up[n].ID)
		return v.up[n].ID, true
	case i > 0:
		return rows[i-1].ID, true
	}
	return "", false
}

// branch draws an outline's part n on its row k, on ground bg, after the
// row's number: a guide for each part it is under, a quiet │ beneath that
// part's caret, then its own caret when it holds rows, ▸ folded and ▾
// unfolded, and the part as its kind draws (part); on the cursor's row,
// its lead too, soft, where it fits before column end, so the rows read as
// names and say more where you are.
func (v *Term) branch(k *ink, n *outline.Node, bg string, on bool, end int) {
	caret := "  "
	if len(held(n)) > 0 {
		caret = "▸ "
		if v.unfolded[n.ID] {
			caret = "▾ "
		}
	}
	k.add(strings.Repeat("│ ", v.depth(n)), tty.Quiet, bg, false)
	k.add(caret, tty.Accent, bg, false)
	v.part(k, n, bg, on)
	if l := lead(n); l != nil && on && n.Kind != "row" && end-k.w-2 >= 20 {
		k.add("  "+printable(l.Text), tty.Soft, bg, false)
	}
}

// part draws a part on a row by its kind: a title bold; a row's name in a
// column as wide as its fellows', its lead beside; an item's or a step's
// sign before its text; a code's first line on the panel; a paragraph
// soft, as what is said is; and any other kind in its look (look), a
// section in the accent and bold, as a head is, a declaration's name
// bold, as a man page has it.
func (v *Term) part(k *ink, n *outline.Node, bg string, on bool) {
	text := name(n)
	switch n.Kind {
	case "title":
		k.add(text, tty.Fore, bg, true)
	case "row":
		k.add(text, tty.Fore, bg, on)
		k.pad(k.w+column(v.up[n].Kids)-tty.Width(text), bg)
		if l := lead(n); l != nil {
			k.add("  "+printable(l.Text), tty.Soft, bg, false)
		}
	case "item", "step":
		s, c := sign(n)
		k.add(s, c, bg, false).add(text, tty.Fore, bg, on)
	case "code":
		more := ""
		if strings.Contains(n.Text, "\n") {
			more = " …"
		}
		k.add(" "+text+more+" ", tty.Fore, tty.Panel, false)
	case "para":
		k.add(text, tty.Soft, bg, false)
	default:
		l, head := look(n.Kind), n.Kind == "section"
		runs := n.Runs
		if len(runs) == 0 {
			runs = []*outline.Node{{Text: text}}
		}
		for _, r := range runs {
			bold := on || head || r.Kind == "name"
			k.add(printable(r.Text), l.Color, bg, bold)
		}
		k.add(l.Mark, tty.Faint, bg, false)
	}
}

// name is a part's text as a row shows it: its first line, printable.
func name(n *outline.Node) string {
	first, _, _ := strings.Cut(n.Text, "\n")
	return printable(first)
}

// tally is what an outline's part holds, said at its row's right: a
// code's lines, a list's steps done, or how many rows it holds, named by
// their kind when they share one.
func tally(n *outline.Node) string {
	if n.Kind == "code" {
		return count(strings.Count(n.Text, "\n")+1, "line")
	}
	rows := held(n)
	if len(rows) == 0 {
		return ""
	}
	kind, done := rows[0].Kind, 0
	for _, r := range rows {
		if r.Kind != kind {
			kind = "part"
		}
		if slices.Contains(r.Tags, "done") {
			done++
		}
	}
	if kind == "step" {
		return fmt.Sprintf("%d of %d done", done, len(rows))
	}
	return count(len(rows), kind)
}

// count is n things, said as a number and a word, the word plural but
// for one.
func count(n int, word string) string {
	if n != 1 {
		word += "s"
	}
	return fmt.Sprintf("%d %s", n, word)
}
