package main

import (
	"fmt"
	"os"
	"strings"

	"sys/outline"
	"sys/tty"
)

const saidAt = 46 // sys/view's: where a thing's facts stand

var marks = map[string]string{"dir": "/", "link": "@", "exec": "*", "bad": "@"}

// -- a line, as sys/view's strip

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

// crumbs is the head: the path from home in its parts, the last bold, or,
// given a label, the label bold after the file
func crumbs(path, label string) *strip {
	parts := strings.Split(strings.TrimPrefix(home(path), "/"), "/")
	last, sep := label, " : "
	if label == "" {
		n := len(parts) - 1
		last, parts, sep = parts[n], parts[:n], " / "
	}
	k := (&strip{}).add(" ", "", "", false)
	for i, p := range parts {
		k.add(p, tty.Soft, "", false)
		if i == len(parts)-1 {
			k.add(sep, tty.Quiet, "", false)
		} else {
			k.add(" / ", tty.Quiet, "", false)
		}
	}
	return k.add(last, tty.Fore, "", true)
}

func rule(w int) *strip {
	return (&strip{}).add(
		strings.Repeat("─", max(0, w)),
		tty.Rule,
		"",
		false,
	)
}

// -- fm's pane

type spot struct {
	*strip
	id string // what a click on the line names
}

func (p *pane) list() []spot {
	head := p.head
	if len(p.rows) > 0 {
		head, _ = outline.Address(p.rows[p.cur].n.ID)
	}
	chips := (&strip{}).add(" t ", "", tty.Rule, false).
		add(" layout", tty.Soft, "", false)
	sw := (&strip{}).add(" ", "", "", false)
	for _, name := range []string{"folder", "tree", "thumbs"} {
		on := name == "folder" && !p.asTree ||
			name == "tree" && p.asTree
		c := tty.Quiet
		if on {
			c = tty.Accent
		}
		sw.add(name, c, "", on).add("  ", "", "", false)
	}
	lines := []spot{{strip: crumbs(head, "").right(chips, p.cols)},
		{
			strip: rule(p.cols),
		}, {strip: &strip{}}, {strip: sw}, {strip: &strip{}}}
	room := max(1, p.high-7) // less the head's five and the foot's two
	n := room
	if len(p.rows) > n {
		n--
	}
	p.page = max(1, n)
	off := max(0, min(p.cur-n/2, len(p.rows)-n))
	for i := off; i < min(off+n, len(p.rows)); i++ {
		lines = append(lines, spot{p.row(i), p.rows[i].n.ID})
	}
	if len(p.rows) > n {
		more := fmt.Sprintf(" +%d", len(p.rows)-n)
		lines = append(
			lines,
			spot{
				strip: (&strip{}).add(
					more,
					tty.Quiet,
					"",
					false,
				),
			},
		)
	}
	for len(lines) < p.high-1 {
		lines = append(lines, spot{strip: &strip{}})
	}
	at := (&strip{}).add(
		strings.Join(p.state, " ")+"  ",
		tty.Faint,
		"",
		false,
	)
	if len(p.rows) > 0 {
		at.add(
			fmt.Sprintf("%d/%d", p.cur+1, len(p.rows)),
			tty.Soft,
			"",
			false,
		)
	}
	foot := (&strip{}).add(" "+p.status, "", "", false).right(at, p.cols)
	return append(lines, spot{strip: foot})
}

// row is sys/view's paint: edge, number, lead and hint, name and mark, and
// the facts where they fit whole
func (p *pane) row(i int) *strip {
	r := p.rows[i]
	on := i == p.cur
	bg, edge, num := "", " ", tty.Quiet
	if on {
		bg, edge, num = tty.Bar, "▌", tty.Soft
	} else if r.n.Sel {
		bg = tty.Rule
	}
	hint := " "
	if r.n.Kind == "dir" || len(r.n.Kids) > 0 {
		hint = "+"
		if r.n.Open {
			hint = "-"
		}
	}
	color, ok := map[string]string{"dir": tty.Accent, "link": tty.Link,
		"exec": tty.Plus, "bad": tty.Minus}[r.n.Kind]
	if !ok {
		color = tty.Fore
	}
	k := (&strip{}).add(edge, tty.Accent, bg, false)
	k.add(fmt.Sprintf("%2d  ", i+1), num, bg, false)
	k.add(strings.Repeat("  ", r.depth)+hint+" ", tty.Faint, bg, false)
	k.add(r.n.Text, color, bg, on)
	k.add(marks[r.n.Kind], tty.Faint, bg, false)
	said := strings.Join(r.n.Tags, " ")
	if at := max(saidAt, k.w+2); said != "" &&
		at+len([]rune(said)) <= p.cols-1 {
		k.pad(at, bg).add(said, tty.Faint, bg, false)
	}
	return k.pad(p.cols-1, bg)
}

// -- ed's pane

func indent(s string) int {
	t := strings.TrimLeft(s, " \t")
	if t == "" {
		return 0
	}
	lead := s[:len(s)-len(t)]
	return len(lead) + 3*strings.Count(lead, "\t")
}

func (p *pane) source() []spot {
	if p.tree == nil || len(p.rows) == 0 {
		return nil
	}
	cur := p.cur
	text := p.rows[cur].n.Text
	label, _, _ := strings.Cut(strings.TrimSpace(text), "  ")
	// the cursor's item: the lines under it that stand further in
	end := cur
	for end+1 < len(p.rows) &&
		indent(p.rows[end+1].n.Text) > indent(text) {
		end++
	}
	span := fmt.Sprintf("%s:%d", home(p.tree.Text), cur+1)
	if end > cur {
		span += fmt.Sprintf("..%d", end+1)
	}
	written := (&strip{}).add(" as written", tty.Soft, "", false).
		add(" · ", tty.Quiet, "", false).
		add(span, tty.Soft, "", false)
	lines := []spot{
		{strip: crumbs(p.tree.Text, label)},
		{strip: rule(p.cols)},
		{strip: &strip{}},
		{strip: written},
		{strip: &strip{}},
	}
	body := max(1, p.high-6)
	p.page = body
	p.first = max(0, min(p.first, cur))
	if cur >= p.first+body {
		p.first = cur - body + 1
	}
	col := p.column(cur)
	wide := max(1, p.cols-6)
	p.off = max(0, min(p.off, col))
	if col >= p.off+wide {
		p.off = col - wide + 1
	}
	for i := p.first; i < p.first+body; i++ {
		if i < len(p.rows) {
			lines = append(
				lines,
				spot{strip: p.line(i, i >= cur && i <= end)},
			)
		} else {
			lines = append(lines, spot{strip: &strip{}})
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
		fmt.Sprintf("%d:%d", cur+1, p.runeAt(cur)+1),
		tty.Soft,
		"",
		false,
	)
	return append(lines, spot{strip: foot.right(at, p.cols)})
}

// runeAt is the cursor's rune in line i; column is where that shows
func (p *pane) runeAt(i int) int {
	n := 0
	for _, r := range p.rows[i].n.Kids {
		if r.Cur {
			break
		}
		n += len([]rune(r.Text))
	}
	return n
}

func (p *pane) column(i int) int {
	col := 0
	for j, c := range []rune(p.rows[i].n.Text) {
		if j == p.runeAt(i) {
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

// line is ed's paintLine: the number, then the text, tabs to their stops,
// the selection and the cursor on grounds of their own
func (p *pane) line(i int, lit bool) *strip {
	n := p.rows[i].n
	bg, num := "", tty.Quiet
	if lit {
		bg, num = tty.Bar, tty.Soft
	}
	k := (&strip{}).add(fmt.Sprintf(" %3d  ", i+1), num, bg, false)
	runs := n.Kids
	if len(runs) == 0 {
		runs = []*outline.Node{{Text: n.Text}}
	}
	insert := len(p.state) > 0 && p.state[0] == "insert"
	// the label of the cursor's line is bold: from its first rune to two
	// blanks
	from := len(
		[]rune(n.Text),
	) - len(
		[]rune(strings.TrimLeft(n.Text, " \t")),
	)
	head, _, _ := strings.Cut(strings.TrimLeft(n.Text, " \t"), "  ")
	to := from + len([]rune(head))
	vis, at := 0, 0
	put := func(c rune, sel, cursor bool) {
		s := string(c)
		if c == '\t' {
			s = strings.Repeat(" ", 4-vis%4)
		} else if c == '\n' {
			s = " "
		}
		for _, one := range s {
			if vis >= p.off && vis-p.off < p.cols-6 {
				fg, g := tty.Fore, bg
				if sel {
					g = tty.Quiet
				}
				if cursor && !insert {
					fg, g = "#121212", "#f8f8f2"
				}
				bold := i == p.cur && at >= from && at < to
				if cursor &&
					insert { // underlined, so alone
					k.runs = append(
						k.runs,
						run{
							string(one),
							fg,
							g,
							bold,
							true,
						},
					)
					k.w++
				} else {
					k.add(string(one), fg, g, bold)
				}
			}
			vis++
			cursor = false
		}
		at++
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
	return k.pad(p.cols, bg)
}
