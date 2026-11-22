package outline

import (
	"bufio"
	"fmt"
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
// address too. Nothing is guessed past the indent: text with no structure
// is a list of lines, and is made structured by indenting it.
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
		at := fmt.Sprintf("%s:%d", id, ln)
		n := &Node{ID: at, Kind: "item", Text: body}
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
// dash, with a box after it or not ("- [ ] "); or a number or a label,
// all up to the last run of two spaces or more ("1  ", "view     ").
func mark(s string) int {
	if strings.HasPrefix(s, "- ") {
		n := 2
		box := s[n:min(len(s), n+4)]
		if box == "[ ] " || box == "[x] " {
			n += 4
		}
		return n + gap(s[n:])
	}
	at := 0
	for i := 0; i < len(s); {
		if s[i] != ' ' {
			i++
			continue
		}
		g := gap(s[i:])
		if g >= 2 && i+g < len(s) {
			at = i + g
		}
		i += g
	}
	return at
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
