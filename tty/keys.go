package tty

import (
	"fmt"
	"strings"
)

// Help paints a titled key/description grid and waits for a key. Rows with
// an empty desc are dropped and " " shows as "space"; the grid is packed
// into as many columns as the terminal height needs, so a long keymap never
// scrolls off the top.
func Help(t *Term, title string, keys [][2]string) {
	type cell struct{ key, desc string }
	var cs []cell
	keyw := 0
	for _, k := range keys {
		key, desc := k[0], k[1]
		if desc == "" {
			continue
		}
		if key == " " {
			key = "space"
		}
		if len(key) > keyw {
			keyw = len(key)
		}
		cs = append(cs, cell{key, desc})
	}
	_, h := t.Size()
	rows := h - 4 // title, a blank, a blank, "press any key"
	if rows < 1 {
		rows = 1
	}
	cols := (len(cs) + rows - 1) / rows
	rows = (len(cs) + cols - 1) / cols // balance a short last column
	colw := make([]int, cols)          // widest desc per column
	for i, c := range cs {
		if n := len([]rune(c.desc)); n > colw[i/rows] {
			colw[i/rows] = n
		}
	}
	t.Write(Home + Clear + Bold(title) + "\r\n\r\n")
	for r := 0; r < rows; r++ {
		line := ""
		for c := 0; c < cols && c*rows+r < len(cs); c++ {
			x := cs[c*rows+r]
			key := Yellow(fmt.Sprintf("%-*s", keyw, x.key))
			line += "  " + key + " " + x.desc
			if (c+1)*rows+r < len(cs) { // next col follows
				pad := colw[c] - len([]rune(x.desc)) + 2
				line += strings.Repeat(" ", pad)
			}
		}
		t.Write(line + "\r\n")
	}
	t.Write(Dim("\r\n  press any key"))
	t.Flush()
	t.ReadKey()
}
