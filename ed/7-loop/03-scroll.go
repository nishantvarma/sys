// scroll keeps the cursor within a rows-tall, cols-wide window. top and off
// are the window's first visible line and visual column.
func (e *editor) scroll(rows, cols int) {
	h := e.head()
	switch {
	case h.line < e.top:
		e.top = h.line
	case h.line >= e.top+rows:
		e.top = h.line - rows + 1
	}
	if e.top < 0 {
		e.top = 0
	}
	ln := e.b.lines[h.line]
	v := visualCol(ln, h.col)
	switch {
	case v < e.off:
		e.off = v
	case v >= e.off+cols:
		e.off = v - cols + 1
	}
	// top can hold its ground because lines follow it down the screen. A
	// short line has nothing to its right, so off has to give ground back.
	if end := visualCol(ln, len(ln)); end < e.off+cols {
		e.off = max(0, end-cols+1)
	}
}
