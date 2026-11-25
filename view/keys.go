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

// Handle is a key given to the view, as a host gives one, rather than read by
// it. What is the view's own it does itself. In a figure, down and up scroll
// it, a line or a page, and g and G to its top and its bottom. Elsewhere, t
// turns to the other layout, the folder's rows or thumbs; a move goes down,
// up, to the last row or a page on, and as thumbs right and left too. The
// program hears where the cursor landed, and every other key, a digit too,
// for a row's number takes the keys after it, and only Key reads keys; false
// when there is nothing for it to hear.
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
		v.lay = (v.lay + 1) % len(layouts)
		v.show(v.status)
		return outline.Key{}, false
	}
	if to, moved := v.move(name); moved {
		return outline.Key{Node: to}, true
	}
	return theirs, true
}

// Ask reads a line at the bottom of the whole terminal, a region's view's
// too, until studio hosts the slots and gives each view its own
// (`todo/03-studio/01-slots`). Completion stays tty's
// seam, which fm leaves empty.
func (v *Term) Ask(prompt, def string) (string, bool) {
	return tty.ReadLine(v.t, prompt, nil, def)
}

// scroll does a key that scrolls a figure: down and up a line or a page,
// g to its top and G to its bottom, where drawing it stops; false when
// the key is no scroll.
func (v *Term) scroll(name string) bool {
	_, h := v.Size()
	top, foot := chrome(h)
	page := max(1, h-top-foot-1)
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
// thumbs, down and up go a row of cards, and right and left one card.
func (v *Term) move(name string) (to string, moved bool) {
	if v.last == nil {
		return "", false
	}
	rows := v.rows()
	w, h := v.Size()
	top, foot := chrome(h)
	page := max(1, h-top-foot-1)
	step := 1
	if v.layout() == "thumbs" {
		step = perRow(w)
		page = max(1, (h-top-foot+1)/3)
	}
	d := 0
	switch name {
	case "j", tty.KDown:
		d = step
	case "k", tty.KUp:
		d = -step
	case tty.KPgDn:
		d = page * step
	case tty.KPgUp:
		d = -page * step
	case "G":
		d = len(rows)
	case tty.KRight, tty.KLeft:
		if v.layout() != "thumbs" {
			return "", false // in and out: the program's
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
	return rows[max(0, min(curRow(rows)+d, len(rows)-1))].ID, true
}

// jump reads a row's number, the digits so far shown as the status. It
// commits once no longer number could stay in range, or at Enter, or at
// any other key, which is then read again as itself; Esc gives up, and a
// resize goes on reading. It is the id of the row the number names, else
// empty.
func (v *Term) jump(first string) string {
	buf := first
	rows := v.rows()
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
		return rows[x-1].ID
	}
	return ""
}
