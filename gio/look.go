package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sys/outline"
	"sys/tty"
	"sys/view"
)

// -- a line, as the window draws it: runs of text, a rune a column

// run is text in one look
type run struct {
	text            string
	fg, bg          string // #rrggbb; no bg is the ground's
	bold, underline bool
}

type strip struct {
	runs []run
	w    int
}

// add puts t at the end of the line, in fg on bg, "" for the ground's;
// neighbours in one look are one run
func (k *strip) add(t, fg, bg string, bold bool) *strip {
	if t == "" {
		return k
	}
	if fg == "" {
		fg = tty.Fore
	}
	k.w += len([]rune(t))
	if n := len(k.runs); n > 0 {
		if r := &k.runs[n-1]; r.fg == fg && r.bg == bg &&
			r.bold == bold &&
			!r.underline {
			r.text += t
			return k
		}
	}
	k.runs = append(k.runs, run{text: t, fg: fg, bg: bg, bold: bold})
	return k
}

func (k *strip) pad(to int, bg string) *strip {
	return k.add(strings.Repeat(" ", max(0, to-k.w)), "", bg, false)
}

// right puts r at the line's right end, w wide, when it fits whole
func (k *strip) right(r *strip, w int) *strip {
	if k.w+r.w < w {
		k.pad(w-r.w-1, "")
		k.runs, k.w = append(k.runs, r.runs...), k.w+r.w
	}
	return k
}

func home(p string) string {
	if h, err := os.UserHomeDir(); err == nil &&
		strings.HasPrefix(p, h+"/") {
		return p[len(h)+1:]
	}
	return p
}

// -- what is no text: sys/view's lines, drawn here

// list is the pane as sys/view lays it out, in the region the pane is,
// as on the terminal: the head, the layouts, the rows or the cards, the
// bottom line; a program with lenses has its bar in place of the head,
// as ed's terminal view puts it (`ed/4-view/34-lensed.go`)
func (p *pane) list() []*strip {
	var lines []*strip
	for _, rs := range p.v.Lines() {
		lines = append(lines, stripOf(rs))
	}
	if p.tops() > 0 && len(lines) > 0 {
		lines[0], _ = p.bar()
	}
	return lines
}

// stripOf is a line of sys/view's runs as the window draws it
func stripOf(rs []view.Run) *strip {
	k := &strip{}
	for _, r := range rs {
		k.add(r.Text, r.Color, r.Ground, r.Bold)
	}
	return k
}

// -- the lenses, a program's ways to see its thing, which it names on its
// tree's tags, ed's (`design/lens`)

// tops is the lines a pane keeps over a text for its program's lenses:
// the bar and its rule, as sys/view gives them (view.Tops); none for a
// program with no lenses
func (p *pane) tops() int {
	if p.tree == nil || len(p.tree.Tags) == 0 {
		return 0
	}
	return view.Tops(p.high)
}

// bar is the pane's top line, its program's lenses as sys/view lays them
// out, the one its state says lit, and where each chip stands
func (p *pane) bar() (*strip, []view.Chip) {
	names := p.tree.Tags
	runs, at := view.Bar(names, view.Shown(names, p.state), p.cols)
	return stripOf(runs), at
}

// barAt is the lens a press at column x of the bar names, "" for none
func (p *pane) barAt(x int) string {
	_, at := p.bar()
	for _, c := range at {
		if c.From <= x && x < c.To {
			return c.Lens
		}
	}
	return ""
}

// -- a text, ed's: its lines after their numbers, and its status

// gutter is ed's: a margin, a line's number in three digits or as many as
// the last needs, and two blanks (`ed/4-view/01-look.go`)
func gutter(n int) int { return max(3, len(strconv.Itoa(n))) + 3 }

// source is ed's pane, as ed draws it on the terminal: the bar and its
// rule on top, then a line a row, and the status at the bottom
func (p *pane) source() []*strip {
	if p.tree == nil || len(p.rows) == 0 {
		return nil
	}
	top := p.tops()
	body := max(1, p.high-1-top)
	p.first = max(0, min(p.first, p.cur))
	if p.cur >= p.first+body {
		p.first = p.cur - body + 1
	}
	g := gutter(len(p.rows))
	col := p.column(p.cur)
	wide := max(1, p.cols-g)
	p.off = max(0, min(p.off, col))
	if col >= p.off+wide {
		p.off = col - wide + 1
	}
	var lines []*strip
	if top > 0 {
		bar, _ := p.bar()
		lines = append(lines, bar)
	}
	if top > 1 {
		lines = append(lines, stripOf(view.Rule(p.cols)))
	}
	for i := p.first; i < p.first+body; i++ {
		if i < len(p.rows) {
			lines = append(lines, p.line(i, g))
		} else {
			lines = append(lines, &strip{})
		}
	}
	mode := ""
	if len(p.state) > 0 {
		mode = p.state[0]
	}
	colours := map[string]string{"insert": tty.Plus, "extend": tty.Link}
	colour := colours[mode]
	if colour == "" {
		colour = tty.Accent
	}
	info, fg := p.tree.Text, tty.Soft
	if short := home(info); short != info {
		info = "~/" + short
	}
	if len(p.state) > 1 {
		info += " " + strings.Join(p.state[1:], " ")
	}
	if p.status != "" {
		info, fg = p.status, tty.Fore
	}
	foot := (&strip{}).add(" "+mode, colour, "", true).
		add("  "+info, fg, "", false)
	at := (&strip{}).add(
		fmt.Sprintf("%d:%d", p.cur+1, p.runeAt(p.cur)+1),
		tty.Soft,
		"",
		false,
	)
	return append(lines, foot.right(at, p.cols))
}

// runeAt is the cursor's rune in line i; column is where that shows
func (p *pane) runeAt(i int) int {
	n := 0
	for _, r := range p.rows[i].Kids {
		if r.Cur {
			break
		}
		n += len([]rune(r.Text))
	}
	return n
}

func (p *pane) column(i int) int {
	col, at := 0, p.runeAt(i)
	for j, c := range []rune(p.rows[i].Text) {
		if j == at {
			break
		}
		if c == '\t' {
			col += 4 - col%4
		} else {
			col++
		}
	}
	return col
}

// cell puts one cell of a text at the end of k: on the selection's
// ground when selected; the cursor's, a block, or in insert underlined,
// a run alone
func cell(k *strip, s string, sel, cursor, insert bool) {
	bg := ""
	if sel {
		bg = tty.Quiet
	}
	switch {
	case cursor && insert:
		k.runs = append(k.runs, run{s, tty.Fore, bg, false, true})
		k.w++
	case cursor:
		k.add(s, groundHex, caretHex, false)
	default:
		k.add(s, tty.Fore, bg, false)
	}
}

// spot is the line and rune a cell of a text shows, x and y in cells
// from the pane's corner: the line's id, and the rune counted from 1, its
// end past the line's end; "" off the lines
func (p *pane) spot(x, y int) (string, int) {
	y -= p.tops()
	i := p.first + y
	if p.tree == nil || y < 0 || y >= p.high-1-p.tops() ||
		i >= len(p.rows) {
		return "", 0
	}
	col := x - gutter(len(p.rows)) + p.off
	r, vis := 0, 0
	for _, c := range []rune(p.rows[i].Text) {
		w := 1
		if c == '\t' {
			w = 4 - vis%4
		}
		if vis+w > col {
			break
		}
		vis += w
		r++
	}
	return p.rows[i].ID, r + 1
}

// line is ed's paintLine: the number, then the text, tabs to their stops,
// the selection and the cursor on grounds of their own
func (p *pane) line(i, g int) *strip {
	n := p.rows[i]
	k := (&strip{}).add(fmt.Sprintf(" %*d  ", g-3, i+1), tty.Quiet, "",
		false)
	runs := n.Kids
	if len(runs) == 0 {
		runs = []*outline.Node{{Text: n.Text}}
	}
	insert := len(p.state) > 0 && p.state[0] == "insert"
	vis := 0
	put := func(c rune, sel, cursor bool) {
		s := string(c)
		if c == '\t' {
			s = strings.Repeat(" ", 4-vis%4)
		} else if c == '\n' {
			s = " "
		}
		for _, one := range s {
			if vis >= p.off && vis-p.off < p.cols-g {
				cell(k, string(one), sel, cursor, insert)
			}
			vis++
			cursor = false
		}
	}
	for _, r := range runs {
		cs := []rune(r.Text)
		if r.Cur && len(cs) == 0 {
			cs = []rune{' '}
		}
		for j, c := range cs {
			put(c, r.Sel, r.Cur && j == 0)
		}
	}
	return k.pad(p.cols, "")
}
