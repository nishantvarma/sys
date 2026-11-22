// scroll keeps the cursor within a rows-tall, cols-wide window. top and off
// are the window's first visible line and visual column.
func (v *term) scroll(n *outline.Node, line, rows, cols int) {
	switch {
	case line < v.top:
		v.top = line
	case line >= v.top+rows:
		v.top = line - rows + 1
	}
	if v.top < 0 {
		v.top = 0
	}
	cur := n.Kids[line]
	ln := []rune(cur.Text)
	x := visualCol(ln, cursorCol(cur))
	switch {
	case x < v.off:
		v.off = x
	case x >= v.off+cols:
		v.off = x - cols + 1
	}
	// top can hold its ground because lines follow it down the screen. A
	// short line has nothing to its right, so off has to give ground back.
	if end := visualCol(ln, len(ln)); end < v.off+cols {
		v.off = max(0, end-cols+1)
	}
}
