package tty

import (
	"fmt"
	"strconv"
	"strings"
)

// Ground is what the views draw on: "", the terminal's own, so that fm
// and ed stand on the same black as the shell beside them. The page's own
// ground is --bg, #111413; a ground the views paint is set here.
const Ground = ""

// The palette the views draw with: the page's tokens in their dark form,
// so that the terminal and the page read as one thing
// (`medium/page.html:/-- tokens --/`), and five a terminal needs that a
// page does not. Each is a colour as the page writes one, #rrggbb, which
// the terminal draws in 24-bit colour, and tmux passes on to a terminal
// that has it.
const (
	Panel  = "#181c1a" // --panel: a card, raised off the ground
	Fore   = "#e3e7e4" // --fg: a name, the loudest thing
	Soft   = "#9aa39e" // --dim: what is said about a thing
	Rule   = "#262b28" // --line: a rule, a border, a key's chip
	Accent = "#6aa2f2" // --accent: the cursor's edge, a folder, the one on
	Plus   = "#52b57c" // --plus: a program, a line added
	Minus  = "#ee665b" // --minus: a link gone, a line removed
	Quiet  = "#4a524e" // a row's number: there, and nearly not
	Faint  = "#6f7873" // marks and a figure's lines: the scaffolding
	Link   = "#5fb8c2" // a link
	Bar    = "#1c2a3d" // the cursor's row: the accent sunk into the ground
	Gone   = "#a3504b" // where a broken link points
)

// Paint is s drawn in colour fg on ground bg, each from the palette, or
// "" for the terminal's own, and bold when asked; the look ends with s.
func Paint(s, fg, bg string, bold bool) string {
	var b strings.Builder
	b.WriteString("\x1b[0")
	if bold {
		b.WriteString(";1")
	}
	if fg != "" {
		b.WriteString(";38;2;" + rgb(fg))
	}
	if bg != "" {
		b.WriteString(";48;2;" + rgb(bg))
	}
	b.WriteString("m" + s + Sgr0)
	return b.String()
}

// rgb is a colour, #rrggbb, as SGR says one: r;g;b.
func rgb(c string) string {
	n, _ := strconv.ParseUint(strings.TrimPrefix(c, "#"), 16, 32)
	return fmt.Sprintf("%d;%d;%d", n>>16, n>>8&0xff, n&0xff)
}
