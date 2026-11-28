package outline

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf8"
)

// Text reads text as an outline, by its indent, as Python reads code: a
// line's kids are the lines indented under it. One line more joins the
// line above instead, as a wrapped item does: a line that hangs under that
// line's text, past its mark (see mark), with no blank line between and no
// mark of its own. A blank line only parts. Tabs count to the next multiple
// of 8. Each node's id is id:n, n the line it starts on, so it is an
// address too, and its kind is line. Nothing is guessed past the indent:
// text with no structure is a list of lines, and is made structured by
// indenting it. It is the floor, what reads a text no reader knows the
// kind of (Parts, Go), and what a picture is read by (sys/picture).
func Text(r io.Reader, id string) []*Node {
	type level struct {
		in int // the column its line starts at
		n  *Node
	}
	var top []*Node
	var up []level
	hang := -1 // the column a wrapped line goes on at; -1, none
	s := bufio.NewScanner(r)
	s.Buffer(nil, 1<<20)
	for ln := 1; s.Scan(); ln++ {
		line := untab(s.Text())
		body := strings.TrimLeft(line, " ")
		if body == "" {
			hang = -1
			continue
		}
		in := len(line) - len(body)
		mk := mark(body)
		body = strings.TrimRight(body, " ")
		if in == hang && mk == 0 {
			n := up[len(up)-1].n
			n.Text += " " + body
			continue
		}
		at := At(id, ln)
		n := &Node{ID: at, Kind: "line", Text: body}
		for len(up) > 0 && up[len(up)-1].in >= in {
			up = up[:len(up)-1]
		}
		if len(up) == 0 {
			top = append(top, n)
		} else {
			p := up[len(up)-1].n
			p.Kids = append(p.Kids, n)
		}
		up = append(up, level{in, n})
		hang = -1
		if mk > 0 {
			hang = in + utf8.RuneCountInString(body[:mk])
		}
	}
	return top
}

// mark is how long the mark that leads a line's text is, 0 with none: a
// dash, with a box after it or not ("- [ ] "), and the spaces after it;
// or a row's name and the run of spaces after it, as Parts reads a row
// (cut): "1  ", "view     ".
func mark(s string) int {
	if m := marked.FindString(s); m != "" {
		return len(m) + gap(s[len(m):])
	}
	if _, said := cut(s); said != "" {
		return len(s) - len(said)
	}
	return 0
}

// gap is how many spaces lead s.
func gap(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

// untab is line with each tab spaced out to the next multiple of 8.
func untab(line string) string {
	if !strings.Contains(line, "\t") {
		return line
	}
	var b strings.Builder
	col := 0
	for _, r := range line {
		if r == '\t' {
			b.WriteString(strings.Repeat(" ", 8-col%8))
			col += 8 - col%8
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String()
}
