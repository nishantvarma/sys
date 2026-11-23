package figure

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"sys/outline"
)

// split parts an item's text at its first run of two spaces or more: the
// label before, the note after with its spaces closed up. An item with no
// such run is a label alone.
func split(text string) (label, note string) {
	label, note, _ = strings.Cut(text, "  ")
	note = strings.Join(strings.Fields(note), " ")
	return strings.TrimSpace(label), note
}

// wrap breaks s at its spaces into lines of at most w runes; a word longer
// than that is a line by itself.
func wrap(s string, w int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		if line != "" && runes(line)+1+runes(word) > w {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// runes is how many cells s takes: one a rune.
func runes(s string) int { return utf8.RuneCountInString(s) }

// widest is the most runes any of lines has.
func widest(lines ...string) int {
	n := 0
	for _, l := range lines {
		n = max(n, runes(l))
	}
	return n
}

// read is the figure a note's items make. Its head, the first item and
// the items under it, is its title and lead (outline.Head). Each other
// item is a box: its label up to the first run of two spaces, its note
// after; the items under it are the boxes inside it. But an item that
// starts with → is a message, from the box it is under to the box whose
// label follows the arrow, and its note is what the message carries. A
// message names one box, or it is an error at its line.
func read(items []*outline.Node) (*Figure, error) {
	if len(items) == 0 {
		return nil, errors.New("nothing to draw")
	}
	title, lead := outline.Head(items)
	f := &Figure{Title: title, Lead: lead}
	named := map[string][]*box{}
	type message struct {
		from      *box
		to, label string
		at        string
	}
	var messages []message
	var walk func(n *outline.Node, up *box) *box
	walk = func(n *outline.Node, up *box) *box {
		label, note := split(n.Text)
		if to, ok := strings.CutPrefix(label, "→"); ok {
			m := message{up, strings.TrimSpace(to), note, n.ID}
			messages = append(messages, m)
			return nil
		}
		b := &box{label: label, note: note, at: n.ID}
		named[label] = append(named[label], b)
		for _, k := range n.Kids {
			if kb := walk(k, b); kb != nil {
				b.kids = append(b.kids, kb)
			}
		}
		return b
	}
	for _, n := range items[1:] {
		if b := walk(n, nil); b != nil {
			f.lanes = append(f.lanes, b)
		}
	}
	for _, m := range messages {
		to := named[m.to]
		why := ""
		switch {
		case m.from == nil:
			why = "under no box"
		case len(to) == 0:
			why = "no box by that name"
		case len(to) > 1:
			why = fmt.Sprintf("%d boxes by that name", len(to))
		}
		if why != "" {
			return nil, fmt.Errorf("%s: → %s: %s", m.at, m.to, why)
		}
		a := &arrow{from: m.from, to: to[0], label: m.label, at: m.at}
		f.arrows = append(f.arrows, a)
	}
	return f, nil
}
