package view

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"sys/figure"
	"sys/outline"
	"sys/tty"
)

// The look's measures, as the samples drew it: what is said about a thing
// stands from column 46 at the least (factsAt), and takes at most 16 of
// the columns at the right, a name at least 10 beside it; a card is 22
// columns wide, 2 from the next.
const (
	saidAt  = 46
	saidW   = 16
	nameMin = 10
	cardW   = 22
	cardGap = 2
)

// layouts are the view's ways to lay a folder out, in the order t turns
// through them: its rows, and thumbs, a card a thing.
var layouts = []string{"folder", "thumbs"}

// ink is a line as it is drawn: its runs, each in its look, and how many
// columns they take.
type ink struct {
	runs []Run
	w    int
}

// Run is text in one look, as the view draws it: its colour and its
// ground, each from tty's palette or "" for the view's own, and bold or
// not. A host that is no terminal draws a line's runs (Lines).
type Run struct {
	Text, Color, Ground string
	Bold                bool
}

// add puts s at the end of the line, in fg on bg.
func (k *ink) add(s, fg, bg string, bold bool) *ink {
	k.runs = append(k.runs, Run{s, fg, bg, bold})
	k.w += tty.Width(s)
	return k
}

// pad fills the line with blanks on bg up to column to.
func (k *ink) pad(to int, bg string) *ink {
	if k.w < to {
		k.add(strings.Repeat(" ", to-k.w), "", bg, false)
	}
	return k
}

// clip cuts the line to n columns, its last one … when it was longer.
func (k *ink) clip(n int) *ink {
	if k.w <= n {
		return k
	}
	var out []Run
	w := 0
	for _, r := range k.runs {
		var b strings.Builder
		for _, c := range r.Text {
			cw := tty.Width(string(c))
			if w+cw > n-1 {
				if n > 0 {
					b.WriteString(tty.Cut)
					w++
				}
				cut := r
				cut.Text = b.String()
				k.runs, k.w = append(out, cut), w
				return k
			}
			b.WriteRune(c)
			w += cw
		}
		r.Text = b.String()
		out = append(out, r)
	}
	k.runs, k.w = out, w
	return k
}

// String is the line in the terminal's words: each run painted, the
// view's own colour the palette's foreground on its ground.
func (k *ink) String() string {
	var b strings.Builder
	for _, r := range k.runs {
		fg, bg := r.Color, r.Ground
		if fg == "" {
			fg = tty.Fore
		}
		if bg == "" {
			bg = tty.Ground
		}
		b.WriteString(tty.Paint(r.Text, fg, bg, r.Bold))
	}
	return b.String()
}

// screen is the whole screen, w by h: all of it but the bottom line, and
// the bottom line, the status, with word on its left. Above the body, the
// head, and, where there is room, a rule and the switch (see chrome); then
// the body, the folder laid out as the view lays it out.
func (v *Term) screen(word string, w, h int) ([]line, line) {
	bottom := line{ink: v.foot(word, w)}
	if v.last == nil {
		return nil, bottom
	}
	top, foot := chrome(h)
	room := max(0, h-top-foot)
	lines := []line{v.head(w)}
	if top > 1 {
		lines = append(lines, rule(w), blank(w), v.switches(w),
			blank(w))
	}
	var body []line
	switch {
	case v.figured():
		body = v.pictured(w, room)
	case v.layout() == "thumbs":
		body = v.cards(w, room)
	default:
		body = v.listed(w, room)
	}
	for len(body) < room {
		body = append(body, blank(w))
	}
	lines = append(lines, body[:room]...)
	if foot > 1 {
		lines = append(lines, blank(w))
	}
	return lines, bottom
}

// chrome is how many lines the view keeps above its body and below it, h
// high: the head, a rule, a blank, the switch and a blank on top, and a
// blank and the status below. A pane too short for that keeps the head
// and the status alone.
func chrome(h int) (top, foot int) {
	if h < 12 {
		return 1, 1
	}
	return 5, 2
}

// figured is whether the tree drawn last is a figure: a note whose kind
// says to draw it as its parts and the messages between them.
func (v *Term) figured() bool {
	return v.last != nil && v.last.Kind == "figure"
}

// where is the node the head names: the cursor's, else the folder shown;
// as thumbs, the folder its cards are; a figure's, the note.
func (v *Term) where() *outline.Node {
	rows := v.rows()
	if v.figured() || len(rows) == 0 || v.layout() == "thumbs" {
		return v.last
	}
	return rows[curRow(rows)]
}

// head is the top line: where the cursor is (where), with its keys at
// the right (headed).
func (v *Term) head(w int) line {
	n := v.where()
	path, at := outline.Address(n.ID)
	label := ""
	if at > 0 {
		label, _, _ = strings.Cut(printable(n.Text), "  ")
	}
	return v.headed(path, n.ID, label, w)
}

// headed is a head line: path from home in its parts, the last of them
// bold, or, given a label, the label bold after the file; cut from the
// left when long, for its end names the place; then the keys at the right
// (chips), when they fit. The last names id, what the cursor is on, and
// each part before it its folder, the one above the next, so that a click
// on a part goes up there; … names the folder above the first shown.
func (v *Term) headed(path, id, label string, w int) line {
	parts, last := v.crumbs(path), label
	if label == "" {
		parts, last = parts[:len(parts)-1], parts[len(parts)-1]
		path = filepath.Dir(path)
	}
	k, xs := crumb(parts, last, label != "")
	for cut := 1; k.w > w-1 && cut <= len(parts); cut++ {
		k, xs = crumb(append([]string{"…"}, parts[cut:]...), last,
			label != "")
	}
	spots := make([]spot, len(xs))
	spots[len(xs)-1] = spot{x: xs[len(xs)-1], id: id}
	for i := len(xs) - 2; i >= 0; i-- {
		spots[i] = spot{x: xs[i], id: path}
		path = filepath.Dir(path)
	}
	spots[0].x = 0 // the margin is the first part's
	if keys, at := v.chips(w - 1 - k.w - 2); keys.w > 0 {
		k.pad(w-1-keys.w, "")
		spots = append(spots, spot{x: k.w})
		for _, s := range at {
			s.x += k.w
			spots = append(spots, s)
		}
		k.runs, k.w = append(k.runs, keys.runs...), k.w+keys.w
	}
	return line{k.clip(w).pad(w, ""), spots}
}

// crumb is the head's path: a margin, each part soft with a quiet / after
// it, : after the file when the last is an item's label, then the last,
// bold; and the column each part starts at, the last's last.
func crumb(parts []string, last string, item bool) (*ink, []int) {
	k := (&ink{}).add(" ", "", "", false)
	var xs []int
	for i, p := range parts {
		sep := " / "
		if item && i == len(parts)-1 {
			sep = " : "
		}
		xs = append(xs, k.w)
		k.add(printable(p), tty.Soft, "", false)
		k.add(sep, tty.Quiet, "", false)
	}
	xs = append(xs, k.w)
	return k.add(printable(last), tty.Fore, "", true), xs
}

// crumbs is a path as the head says it, in its parts: from home when it is
// under home, ~ for home itself, else from the root, / before the first.
func (v *Term) crumbs(p string) []string {
	h := v.home
	switch {
	case h != "" && p == h:
		return []string{"~"}
	case h != "" && strings.HasPrefix(p, h+"/"):
		return strings.Split(p[len(h)+1:], "/")
	case p == "/":
		return []string{"/"}
	}
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	if strings.HasPrefix(p, "/") {
		parts[0] = "/" + parts[0]
	}
	return parts
}

// chips is the keys the head shows, at most room columns of them: the
// program's (Chips), then the view's own, t for the layout, while the
// view lays out a folder. Each key stands in a chip and what it does after
// it; a range, 1..4, is its two ends in chips. Where each chip stands,
// from the chips' first column, is a spot that presses its key, but a
// range, which is no one key.
func (v *Term) chips(room int) (*ink, []spot) {
	keys := v.keys
	if v.last != nil && !v.figured() {
		own := [2]string{"t", "layout"}
		keys = append(keys[:len(keys):len(keys)], own)
	}
	k := &ink{}
	var at []spot
	for i, c := range keys {
		next := &ink{}
		if i > 0 {
			next.add("   ", "", "", false)
		}
		a, b, ranged := strings.Cut(c[0], "..")
		start := next.w
		next.add(" "+a+" ", tty.Fore, tty.Rule, false)
		if ranged {
			next.add("..", tty.Quiet, "", false)
			next.add(" "+b+" ", tty.Fore, tty.Rule, false)
		}
		next.add(" "+c[1], tty.Soft, "", false)
		if k.w+next.w > room {
			break
		}
		if !ranged {
			at = append(at, spot{x: k.w + start, key: c[0]})
		}
		at = append(at, spot{x: k.w + next.w})
		k.runs, k.w = append(k.runs, next.runs...), k.w+next.w
	}
	return k, at
}

// switches is the line over the body: the program's tabs, when it has
// some (Tabs), else the view's layouts, the one in use lit and the other
// a click's to turn to. A tab is its
// name, in a chip while it is the one shown and dim while it waits to be
// built, then its tally, a count or a diff's +2 −1, and its number.
func (v *Term) switches(w int) line {
	k := (&ink{}).add(" ", "", "", false)
	var spots []spot
	switch {
	case v.tabs != nil:
		for i, t := range v.tabs {
			fg, bg := tty.Soft, ""
			switch {
			case t.On:
				fg, bg = tty.Fore, tty.Rule
			case t.Waits:
				fg = tty.Quiet
			}
			k.add(" "+printable(t.Name)+" ", fg, bg, t.On)
			tally := strings.Fields(printable(t.Tally))
			for _, word := range tally {
				c := tty.Soft
				switch {
				case strings.HasPrefix(word, "+"):
					c = tty.Plus
				case strings.HasPrefix(word, "−"),
					strings.HasPrefix(word, "-"):
					c = tty.Minus
				}
				k.add(word+" ", c, "", false)
			}
			k.add(strconv.Itoa(i+1), tty.Quiet, "", false)
			k.add("  ", "", "", false)
		}
	case !v.figured():
		for _, name := range layouts {
			on := name == v.layout()
			c, key := tty.Quiet, "t"
			if on {
				c, key = tty.Accent, ""
			}
			spots = append(spots, spot{x: k.w, key: key})
			k.add(name, c, "", on)
			spots = append(spots, spot{x: k.w})
			k.add("  ", "", "", false)
		}
	}
	return line{k.clip(w).pad(w, ""), spots}
}

// layout is the name of the layout in use: the folder until t turns it.
func (v *Term) layout() string {
	return layouts[v.lay]
}

// listed is the body as the folder's rows: as many as fit in room lines,
// the cursor's kept in the middle when it can be, and a count of the rows
// that did not fit.
func (v *Term) listed(w, room int) []line {
	rows := v.rows()
	n := room
	if len(rows) > n {
		n--
	}
	n = max(0, n)
	off := max(0, min(curRow(rows)-n/2, len(rows)-n))
	vis := rows[off:min(off+n, len(rows))]
	at := v.factsAt(rows, w)
	var lines []line
	for i, r := range vis {
		lines = append(lines, line{v.paint(off+i, r, w, at),
			[]spot{{x: 0, id: r.ID}}})
	}
	if len(rows) > n && room > 0 {
		lines = append(lines, more(len(rows)-n, w))
	}
	return lines
}

// factsAt is the column the facts of a folder's things stand at, w wide:
// past its longest name that has facts, so that no name is cut while the
// view has room for it, and from saidAt at the least, so that facts line
// up; but left enough for them to fit, the names then cut short before
// them. 0 when the view is too narrow for a name and its facts both, and
// shows the names alone.
func (v *Term) factsAt(rows []*outline.Node, w int) int {
	lead := 1 + tty.Width(fmt.Sprintf("%2d  ", len(rows)))
	name, said := 0, 0
	for _, n := range rows {
		if s := printable(strings.Join(n.Tags, " ")); s != "" {
			shown := printable(n.Text) + v.look(n.Kind).Mark
			name = max(name, tty.Width(shown))
			said = max(said, min(tty.Width(s), saidW))
		}
	}
	at := min(max(saidAt, lead+name+2), w-1-said)
	if at < lead+nameMin+2 {
		return 0
	}
	return at
}

// paint is row n as line i, w wide: the cursor's edge, the row's number,
// then its name in its kind's look and its facts from column at, the name
// cut short before them when it is longer; no facts at 0. The cursor's
// row is lit across, its name bold; numbers are quiet and marks faint,
// so that names are the loudest thing.
func (v *Term) paint(i int, n *outline.Node, w, at int) *ink {
	on := n.Cur
	bg, edge, num := "", " ", tty.Quiet
	switch {
	case on:
		bg, edge, num = tty.Bar, "▌", tty.Soft
	case n.Sel:
		bg = tty.Rule
	}
	look := v.look(n.Kind)
	k := &ink{}
	k.add(edge, tty.Accent, bg, false)
	k.add(fmt.Sprintf("%2d  ", i+1), num, bg, false)
	k.add(printable(n.Text), look.Color, bg, on)
	k.add(look.Mark, tty.Faint, bg, false)
	if said := printable(strings.Join(n.Tags, " ")); said != "" && at > 0 {
		k.clip(at-2).pad(at, bg).add(said, look.Said, bg, false)
	}
	return k.clip(w-1).pad(w, bg)
}

// cards is the body as thumbs: the folder's things as cards, as many
// across as fit (perRow), each its name in its kind's look, bold, and its
// mark, and under it its kind's word and what is said of it; the cursor's
// card lit, with its edge, the others raised. Rows of cards, a blank
// between, keep the cursor's in the middle when they can, and a count
// says how many cards did not fit.
func (v *Term) cards(w, room int) []line {
	rows := v.rows()
	per := perRow(w)
	total := (len(rows) + per - 1) / per
	n := (room + 1) / 3
	if total > n {
		n = room / 3
	}
	n = max(0, n)
	off := max(0, min(curRow(rows)/per-n/2, total-n))
	var lines []line
	for s := off; s < min(off+n, total); s++ {
		a, b, spots := v.card(rows, s*per, per, w)
		lines = append(lines, line{a, spots}, line{b, spots}, blank(w))
	}
	if total > n && room > 0 {
		seen := min(len(rows), (off+n)*per) - off*per
		lines = append(lines[:min(len(lines), room-1)],
			more(len(rows)-seen, w))
	}
	return lines
}

// perRow is how many cards stand across w: a margin, then each card and
// the gap after it, the last gap past the edge.
func perRow(w int) int {
	return max(1, (w-1+cardGap)/(cardW+cardGap))
}

// card is a row of cards, the rows from start, per of them, as its two
// lines and where each card stands on them.
func (v *Term) card(rows []*outline.Node, start, per, w int) (*ink, *ink,
	[]spot) {
	a := (&ink{}).add(" ", "", "", false)
	b := (&ink{}).add(" ", "", "", false)
	var spots []spot
	for i := start; i < min(start+per, len(rows)); i++ {
		n := rows[i]
		bg, edge := tty.Panel, " "
		switch {
		case n.Cur:
			bg, edge = tty.Bar, "▌"
		case n.Sel:
			bg = tty.Rule
		}
		look := v.look(n.Kind)
		name := printable(n.Text)
		said := strings.Join(strings.Fields(printable(
			strings.Join(n.Tags, " "))), " ")
		word := look.Word
		if word == "" {
			word = n.Kind
		}
		a.add(edge, tty.Accent, bg, false)
		a.add(name, look.Color, bg, true)
		a.add(look.Mark, tty.Faint, bg, false)
		b.add(" ", "", bg, false).add(word, tty.Soft, bg, false)
		if said != "" {
			b.add("  "+said, look.Said, bg, false)
		}
		x := 1 + (i-start)*(cardW+cardGap)
		end := min(x+cardW, w)
		for _, l := range []*ink{a, b} {
			l.clip(end).pad(end, bg).pad(min(end+cardGap, w), "")
		}
		spots = append(spots, spot{x: x, id: n.ID},
			spot{x: x + cardW})
	}
	return a.clip(w).pad(w, ""), b.clip(w).pad(w, ""), spots
}

// pictured is the body as a figure, a note whose kind is figure: its
// items laid out as parts and messages (sys/figure) and drawn in runes,
// from the line the view has scrolled to, as many as fit, with a count of
// those below; a note that makes no figure says why. A cell of the
// picture names what the figure says stands there.
func (v *Term) pictured(w, room int) []line {
	f, err := figure.New(v.last.Kids)
	var pic []string
	if err != nil {
		pic = []string{printable(err.Error())}
	} else {
		for _, p := range f.Runes() {
			pic = append(pic, printable(p))
		}
	}
	v.top = max(0, min(v.top, len(pic)-room))
	vis := pic[v.top:min(len(pic), v.top+room)]
	below := len(pic) - v.top - len(vis)
	if below > 0 && len(vis) > 0 {
		vis, below = vis[:len(vis)-1], below+1
	}
	var lines []line
	for i, p := range vis {
		l := line{ink: drawing(" "+p, err != nil).clip(w).pad(w, "")}
		if f != nil {
			l.spots = parts(f, v.top+i, w)
		}
		lines = append(lines, l)
	}
	if below > 0 {
		lines = append(lines, more(below, w))
	}
	return lines
}

// drawing is a line of a figure in its looks: a box's border and a
// message's line faint, a message's head in the accent, words as words;
// a note that made no figure, its why, faint whole.
func drawing(s string, why bool) *ink {
	k := &ink{}
	for _, c := range s {
		fg := tty.Fore
		switch {
		case why, strings.ContainsRune("─│╭╮╰╯┬┴├┤┼┌┐└┘", c):
			fg = tty.Faint
		case strings.ContainsRune("▼▲▶◀", c):
			fg = tty.Accent
		}
		if n := len(k.runs); n > 0 && k.runs[n-1].Color == fg {
			k.runs[n-1].Text += string(c)
			k.w += tty.Width(string(c))
			continue
		}
		k.add(string(c), fg, "", false)
	}
	return k
}

// parts is where the parts of f stand on its line y, drawn from a
// screen's second column, w wide: a spot each time what a cell names
// changes (figure.At).
func parts(f *figure.Figure, y, w int) []spot {
	var spots []spot
	last := ""
	for x := 0; x < w; x++ {
		if id := f.At(x-1, y); id != last {
			spots = append(spots, spot{x: x, id: id})
			last = id
		}
	}
	return spots
}

// more is the line that counts the n that did not fit, quiet.
func more(n, w int) line {
	k := (&ink{}).add(fmt.Sprintf(" +%d", n), tty.Quiet, "", false)
	return line{ink: k.clip(w).pad(w, "")}
}

// blank is an empty line, w wide, on the ground.
func blank(w int) line {
	return line{ink: (&ink{}).pad(w, "")}
}

// rule is a line across, w wide, in the rule's colour.
func rule(w int) line {
	k := (&ink{}).add(strings.Repeat("─", max(0, w)), tty.Rule, "", false)
	return line{ink: k}
}

// look is how the view draws a kind: as the program says, a colour and a
// mark, a word for a card, and a colour for what is said of it; a kind
// the program says nothing of is the foreground's, and what is said of
// it faint.
func (v *Term) look(kind string) Look {
	l := v.looks[kind]
	if l.Color == "" {
		l.Color = tty.Fore
	}
	if l.Said == "" {
		l.Said = tty.Faint
	}
	return l
}

// printable is s as a terminal shows it and nothing more: a tab as a
// space, and any other control character, which would move the cursor or
// change the colours, as ?. A name or a line of a note is drawn so.
func printable(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case unicode.IsControl(r):
			return '?'
		}
		return r
	}, s)
}

// foot is the bottom line, w wide: the program's word on its left, and
// at its right the program's state, faint, then where the cursor is among
// what the view shows, 12/34. Short of room, the word is cut, for the
// state says what holds now.
func (v *Term) foot(word string, w int) *ink {
	right := &ink{}
	if s := printable(strings.Join(v.state, " ")); s != "" {
		right.add(s, tty.Faint, "", false)
	}
	if p := v.pos(); p != "" {
		if right.w > 0 {
			right.add("  ", "", "", false)
		}
		right.add(p, tty.Soft, "", false)
	}
	k := (&ink{}).add(" "+printable(word), tty.Fore, "", false)
	k.clip(max(0, w-right.w-2)).pad(w-right.w-1, "")
	k.runs, k.w = append(k.runs, right.runs...), k.w+right.w
	return k.clip(w).pad(w, "")
}

// pos is where the cursor is among what the view shows, 12/34; nothing
// for a figure, which has no cursor, or for no rows.
func (v *Term) pos() string {
	rows := v.rows()
	if v.figured() || len(rows) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", curRow(rows)+1, len(rows))
}
