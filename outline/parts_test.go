package outline

import (
	"strings"
	"testing"
)

// parted is parts as text again: a part a line, two spaces a level, its
// kind, its text, done for a step done, and its line. A part's text is
// said from its runs when it has them, an address between backticks
// again, and a name it declares between stars.
func parted(b *strings.Builder, ns []*Node, depth int) {
	for _, n := range ns {
		text := n.Text
		if len(n.Runs) > 0 {
			text = ""
			for _, r := range n.Runs {
				switch r.Kind {
				case "addr":
					text += "`" + r.Text + "`"
				case "name":
					text += "*" + r.Text + "*"
				default:
					text += r.Text
				}
			}
		}
		pad := strings.Repeat("  ", depth)
		b.WriteString(pad + n.Kind + " " + text)
		if len(n.Tags) > 0 {
			b.WriteString(" [" + strings.Join(n.Tags, " ") + "]")
		}
		line := n.ID[strings.LastIndex(n.ID, ":"):]
		b.WriteString("   " + line + "\n")
		parted(b, n.Kids, depth+1)
	}
}

// TestParts: one note carried through, each rule in it once: its head; a
// paragraph run on, an address in it; a table under it after a blank
// line, its rows' text hung; a section, a paragraph in it whose later line
// holds a row, which goes on at the paragraph's level; steps, one wrapped
// and holding an item, one done; a fence; and markdown's heads.
func TestParts(t *testing.T) {
	in := "view: a program and its view\n" +
		"\n" +
		"A program says what to show as nodes; a view draws\n" +
		"them, as `view/term.go` does.\n" +
		"\n" +
		"\tsource     the text as written, the one tab\n" +
		"\t           typed in\n" +
		"\toutline    its parts\n" +
		"\n" +
		"Messages, not calls\n" +
		"  A program and its view pass plain data,\n" +
		"  a JSON line each.\n" +
		"    to the view      draw: the tree;\n" +
		"                     ask: a prompt\n" +
		"  A message joins when a view needs one.\n" +
		"\n" +
		"- [ ] a step that\n" +
		"      wraps\n" +
		"  - a kid of the step\n" +
		"- [x] a step done\n" +
		"\n" +
		"```\n" +
		"func f() {\n" +
		"\treturn\n" +
		"}\n" +
		"```\n" +
		"\n" +
		"# a markdown head\n" +
		"a line under it\n" +
		"## a lower head\n" +
		"under that\n" +
		"# another head\n"
	want := "title view: a program and its view   :1\n" +
		"para A program says what to show as nodes; a view draws" +
		" them, as `view/term.go` does.   :3\n" +
		"row source   :6\n" +
		"  para the text as written, the one tab typed in   :6\n" +
		"row outline   :8\n" +
		"  para its parts   :8\n" +
		"section Messages, not calls   :10\n" +
		"  para A program and its view pass plain data, a JSON line" +
		" each.   :11\n" +
		"  row to the view   :13\n" +
		"    para draw: the tree; ask: a prompt   :13\n" +
		"  para A message joins when a view needs one.   :15\n" +
		"step a step that wraps   :17\n" +
		"  item a kid of the step   :19\n" +
		"step a step done [done]   :20\n" +
		"code func f() {\n        return\n}   :22\n" +
		"section a markdown head   :28\n" +
		"  para a line under it   :29\n" +
		"  section a lower head   :30\n" +
		"    para under that   :31\n" +
		"section another head   :32\n"
	ps, err := Parts(strings.NewReader(in), "note")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	parted(&b, ps, 0)
	if b.String() != want {
		t.Errorf("parts:\n%s\nwant:\n%s", b.String(), want)
	}
	want = "A program says what to show as nodes; a view draws them," +
		" as view/term.go does."
	if p := ps[1]; p.Text != want || len(p.Runs) != 3 || p.Kids != nil {
		t.Errorf("a paragraph is its runs, its backticks gone, and no"+
			" parts: %q, %d runs, %d parts", p.Text, len(p.Runs),
			len(p.Kids))
	}
}
