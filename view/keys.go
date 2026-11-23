package view

import (
	"strconv"

	"sys/figure"
	"sys/outline"
	"sys/tty"
)

// Key reads the next key, does what is the view's own (see Handle), and
// hands back the first that is the program's. Outside a figure a digit
// starts a row's number, which takes the keys after it (see jump).
func (v *Term) Key() (outline.Key, bool) {
	for {
		k, ok := v.t.ReadKey()
		if !ok {
			return outline.Key{}, false
		}
		if name := k.Token(); digit(name) && !v.figured() {
			return outline.Key{Node: v.jump(name)}, true
		}
		if ev, theirs := v.Handle(k); theirs {
			return ev, true
		}
	}
}

// Handle is a key given to the view, as a host gives one, rather than
// read by it. What is the view's own it does itself. In a figure, down
// and up scroll it, a line or a page, and g and G to its top and its
// bottom. Elsewhere, t turns to the next layout, the folder, the tree,
// tiles; a move goes down, up, to the last row or a page on, and as tiles
// right and left too; in the tree, a node opens and closes in place. The
// program hears where the cursor landed, what the view wants to see
// inside, and every other key, a digit too, for a row's number takes the
// keys after it, and only Key reads keys; false when there is nothing for
// it to hear.
func (v *Term) Handle(k tty.Key) (outline.Key, bool) {
	name := k.Token()
	theirs := outline.Key{Name: name, Text: k.Text}
	if v.figured() {
		if v.scroll(name) {
			return outline.Key{}, false
		}
		return theirs, true
	}
	if name == "t" {
		next := map[string]string{"": "tree", "tree": "tiles"}
		v.lay = next[v.lay]
		v.show(v.status)
		return outline.Key{}, false
	}
	if to, moved := v.move(name); moved {
		return outline.Key{Node: to}, true
	}
	if ev, turned := v.turn(name); turned {
		return ev, true
	}
	return theirs, true
}

// Ask reads a line at the bottom of the whole terminal, a region's view's
// too, until studio hosts the slots and gives each view its own
// (`todo/studio/outline:/studio hosts the slots/`). Completion stays tty's
// seam, which fm leaves empty.
func (v *Term) Ask(prompt, def string) (string, bool) {
	return tty.ReadLine(v.t, prompt, nil, def)
}

// scroll does a key that scrolls a figure: down and up a line or a page,
// g to its top and G to its bottom, where drawing it stops; false when
// the key is no scroll.
func (v *Term) scroll(name string) bool {
	_, h := v.Size()
	page := max(1, h-3)
	switch name {
	case "j", tty.KDown:
		v.top++
	case "k", tty.KUp:
		v.top--
	case tty.KPgDn:
		v.top += page
	case tty.KPgUp:
		v.top -= page
	case "g":
		v.top = 0
	case "G":
		// past the last page: pictured stops it there
		if f, err := figure.New(v.last.Kids); err == nil {
			v.top = len(f.Runes())
		}
	default:
		return false
	}
	v.top = max(0, v.top)
	v.show(v.status)
	return true
}

// digit is whether a key is a digit, 0 to 9.
func digit(name string) bool {
	return len(name) == 1 && name[0] >= '0' && name[0] <= '9'
}

// move is the id of the row a key moves the cursor to, among those shown,
// empty when there are none; moved is false when the key is no move. As
// tiles, down and up go a line of tiles, and right and left one tile.
func (v *Term) move(name string) (to string, moved bool) {
	if v.last == nil {
		return "", false
	}
	_, rows := v.shown()
	w, h := v.Size()
	step := 1
	if v.lay == "tiles" {
		step, _ = v.grid(rows, w)
	}
	d := 0
	switch name {
	case "j", tty.KDown:
		d = step
	case "k", tty.KUp:
		d = -step
	case tty.KPgDn:
		d = (h - 3) * step
	case tty.KPgUp:
		d = -(h - 3) * step
	case "G":
		d = len(rows)
	case tty.KRight, tty.KLeft:
		if v.lay != "tiles" {
			return "", false // in and out, not a move here
		}
		d = 1
		if name == tty.KLeft {
			d = -1
		}
	default:
		return "", false
	}
	if len(rows) == 0 {
		return "", true
	}
	return rows[max(0, min(curRow(rows)+d, len(rows)-1))].n.ID, true
}

// jump reads a row's number, the digits so far shown as the status. It
// commits once no longer number could stay in range, or at Enter, or at
// any other key, which is then read again as itself; Esc gives up, and a
// resize goes on reading. It is the id of the row the number names, else
// empty.
func (v *Term) jump(first string) string {
	buf := first
	_, rows := v.shown()
	n := len(rows)
	for x, _ := strconv.Atoi(buf); x*10 <= n; x, _ = strconv.Atoi(buf) {
		v.show(buf)
		k, ok := v.t.ReadKey()
		name := k.Token()
		if !ok || name == tty.KEsc || name == tty.KInt {
			return ""
		}
		if name == tty.KResize {
			continue
		}
		if !digit(name) {
			if name != tty.KEnter {
				v.t.Feed([]tty.Key{k})
			}
			break
		}
		buf += name
	}
	if x, err := strconv.Atoi(buf); err == nil && x >= 1 && x <= n {
		return rows[x-1].n.ID
	}
	return ""
}

// turn is going in and out in the tree, the view's own. Enter and right
// open or close a folder or an item with kids in place, and right a file
// too, asking the program for what is inside; enter on a file, or on an
// item with no kids, is the program's, for it edits. Left and h close what
// is open, else go up to the node a row is in. turned is false zoomed,
// where in and out are the program's, and at the top, where going out is.
func (v *Term) turn(name string) (ev outline.Key, turned bool) {
	if v.lay != "tree" || v.last == nil {
		return ev, false
	}
	_, rows := v.shown()
	if len(rows) == 0 {
		return ev, false
	}
	i := curRow(rows)
	r := rows[i]
	switch name {
	case tty.KEnter, tty.KRight:
		leaf := r.n.Kind == "item" && len(r.n.Kids) == 0
		file := r.n.Kind != "dir" && r.n.Kind != "item"
		if leaf || name == tty.KEnter && file {
			return ev, false
		}
		if r.n.Open {
			return outline.Key{Name: "close", Node: r.n.ID}, true
		}
		return outline.Key{Name: "open", Node: r.n.ID}, true
	case tty.KLeft, "h":
		if r.n.Open {
			return outline.Key{Name: "close", Node: r.n.ID}, true
		}
		for j := i - 1; j >= 0; j-- {
			if rows[j].depth < r.depth {
				return outline.Key{Node: rows[j].n.ID}, true
			}
		}
		return ev, false
	}
	return ev, false
}
