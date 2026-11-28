package outline

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Parts reads a note as its writer marks it, a part each, of a kind a view
// draws it by (`design/view:/Kinds, never looks/`); medium/page/list reads
// most of the same marks for the page (`medium/page/list:/def item/`):
//
//	title    the note's first paragraph, its head
//	section  a line with lines indented under it, which are its parts; in
//	         markdown a line of #s, its parts what follows, to the next
//	         line of as many or fewer
//	para     lines run on, to a blank line; what is between backticks in
//	         it an address, a run of kind addr, the text around it runs
//	row      a name, two spaces or more, and what is said of it, the para
//	         under it; a line hung under that text runs on with it
//	item     a line marked by a bullet, "- ", "* " or "+ ", as markdown
//	         has them, its own parts indented under it
//	step     a bullet with a box, "- [ ] ", tagged done when "- [x] "
//	code     the lines between two fences of ```, as written
//
// A line holds what is indented under it only when a part starts on it:
// a paragraph's later line runs on, and a blank line before an indented
// block parts the block from the line above it; either way the block is
// parts of its own, at that line's level, so a table under a paragraph is
// the paragraph's, not its last line's. The note's first part is its
// title, what is indented under it what is said of it. Tabs count to the
// next multiple of 8. Each part's id is id:n, n the line it starts on, so
// it is an address.
func Parts(r io.Reader, id string) ([]*Node, error) {
	var ls []line
	s := bufio.NewScanner(r)
	s.Buffer(nil, 1<<20)
	for n := 1; s.Scan(); n++ {
		t := untab(strings.TrimRight(s.Text(), " \r"))
		body := strings.TrimLeft(t, " ")
		ls = append(ls, line{n, len(t) - len(body), body})
	}
	ps := parts(ls, id)
	spans(ps)
	if len(ps) > 0 && (ps[0].Kind == "para" || ps[0].Kind == "section" &&
		!strings.HasPrefix(ps[0].Text, "#") && head(first(ls)) == 0) {
		ps[0].Kind = "title"
	}
	return ps, s.Err()
}

// line is a line of a note as Parts reads it: its number, from 1, the
// column its text starts at, and its text, "" when it is blank.
type line struct {
	n, in int
	text  string
}

// marked is what marks an item, its first match: a bullet as markdown has
// one, - * or +, and one space, the item's text right after it; and a
// step, its second, a box between, [ ] or [x], and its spaces. A bullet
// two spaces or more before its text is no mark but a row's name, as a
// key table's * is.
var marked = regexp.MustCompile(`^([-*+] (\[[ x]\] +)?)[^ ]`)

// cut parts a row's text at its first run of two spaces or more: its name
// before, and what is said of it after; said is "" when it is no row. Text
// and Parts both read a row so, and sys/picture a box.
func cut(text string) (name, said string) {
	i := strings.Index(text, "  ")
	if i <= 0 {
		return "", ""
	}
	return text[:i], strings.TrimLeft(text[i:], " ")
}

// parts is a block of lines, all at one indent or past it, as its parts,
// each paragraph's text as written: its runs are set after (spans).
func parts(ls []line, id string) []*Node {
	var out []*Node
	gap := true  // a blank line before this one, or the block's start
	run := false // out ends with a paragraph, which a line can run on
	for i := 0; i < len(ls); {
		l := ls[i]
		if l.text == "" {
			gap, i = true, i+1
			continue
		}
		end := i + 1
		switch {
		case fence(l.text):
			for end < len(ls) && !fence(ls[end].text) {
				end++
			}
			out = append(out, code(ls[i+1:end], l.in, At(id, l.n)))
			gap, run, i = false, false, end+1
			continue
		case head(l.text) > 0:
			for end < len(ls) && !(ls[end].in <= l.in &&
				head(ls[end].text) > 0 &&
				head(ls[end].text) <= head(l.text)) {
				end++
			}
			n := &Node{ID: At(id, l.n), Kind: "section",
				Text: strings.TrimLeft(l.text, "# "),
				Kids: parts(ls[i+1:end], id)}
			out = append(out, n)
			gap, run, i = true, false, end
			continue
		}
		for end < len(ls) &&
			(ls[end].text == "" || ls[end].in > l.in) {
			end++
		}
		under := ls[i+1 : end]
		for len(under) > 0 && under[len(under)-1].text == "" {
			under = under[:len(under)-1]
		}
		parted := i+1+len(under) < end // a blank line after the block
		apart := len(under) > 0 && under[0].text == ""
		plain := !marked.MatchString(l.text) && !isRow(l)
		if !gap && run && plain {
			out[len(out)-1].Text += " " + l.text
		} else {
			own := under
			if apart {
				own = nil
			} else {
				under = nil
			}
			n := part(l, own, id)
			out, run = append(out, n), n.Kind == "para"
		}
		// what is under the line, not its own, goes on at its level
		if len(under) > 0 {
			out = append(out, parts(under, id)...)
			parted, run = true, false
		}
		gap, i = parted, end
	}
	return out
}

// part is the part that starts on line l, the lines under it its own: an
// item or a step, a row, a section, or a paragraph's first line.
func part(l line, under []line, id string) *Node {
	at := At(id, l.n)
	if m := marked.FindStringSubmatch(l.text); m != nil {
		text, rest := hung(l.text[len(m[1]):], under, l.in+len(m[1]))
		n := &Node{ID: at, Kind: "item", Text: text,
			Kids: parts(rest, id)}
		if box := m[2]; box != "" {
			n.Kind = "step"
			if box[1] == 'x' {
				n.Tags = []string{"done"}
			}
		}
		return n
	}
	if name, said, col := row(l); said != "" {
		said, rest := hung(said, under, col)
		kids := append([]*Node{{ID: at, Kind: "para", Text: said}},
			parts(rest, id)...)
		return &Node{ID: at, Kind: "row", Text: name, Kids: kids}
	}
	if len(under) > 0 {
		return &Node{ID: at, Kind: "section", Text: l.text,
			Kids: parts(under, id)}
	}
	return &Node{ID: at, Kind: "para", Text: l.text}
}

// row is a line's name and what is said of it (cut), and the column what
// is said starts at.
func row(l line) (name, said string, col int) {
	name, said = cut(l.text)
	return name, said, l.in +
		utf8.RuneCountInString(l.text[:len(l.text)-len(said)])
}

// isRow is whether line l is a row, a name and what is said of it.
func isRow(l line) bool {
	_, said, _ := row(l)
	return said != ""
}

// hung is text run on with the lines hung under it, first in under, at
// column col and with no mark of their own, as a wrapped item's are; and
// the lines under it left.
func hung(text string, under []line, col int) (string, []line) {
	for len(under) > 0 && under[0].in == col && under[0].text != "" &&
		!marked.MatchString(under[0].text) {
		text += " " + under[0].text
		under = under[1:]
	}
	return text, under
}

// fence is whether a line is a fence, ```, which code stands between.
func fence(text string) bool {
	return strings.HasPrefix(text, "```")
}

// first is the text of a note's first line that is not blank.
func first(ls []line) string {
	for _, l := range ls {
		if l.text != "" {
			return l.text
		}
	}
	return ""
}

// head is how high a line of markdown's #s heads what follows, from 1 for
// #, and 0 for a line that is none.
func head(text string) int {
	n := len(text) - len(strings.TrimLeft(text, "#"))
	if n == 0 || n > 6 || !strings.HasPrefix(text[n:], " ") {
		return 0
	}
	return n
}

// code is lines between fences as written, kind code, from the fence's
// column, a line of text each.
func code(ls []line, in int, at string) *Node {
	var out []string
	for _, l := range ls {
		pad := ""
		if l.text != "" {
			pad = strings.Repeat(" ", max(0, l.in-in))
		}
		out = append(out, pad+l.text)
	}
	return &Node{ID: at, Kind: "code", Text: strings.Join(out, "\n")}
}

// spans gives each paragraph its runs (runs).
func spans(ns []*Node) {
	for _, n := range ns {
		if n.Kind == "para" {
			n.Text, n.Runs = runs(n.Text)
		}
		spans(n.Kids)
	}
}

// runs is a paragraph's text as written, its backticks dropped, and its
// runs: what is between a pair of them an address, of kind addr, and the
// text around it runs of no kind; none when it has no pair.
func runs(text string) (string, []*Node) {
	cut := strings.Split(text, "`")
	if len(cut) < 3 {
		return text, nil
	}
	var runs []*Node
	for i, c := range cut {
		if c == "" {
			continue
		}
		kind := ""
		if i%2 == 1 && i < len(cut)-1 {
			kind = "addr"
		}
		runs = append(runs, &Node{Kind: kind, Text: c})
	}
	return strings.Join(cut, ""), runs
}
