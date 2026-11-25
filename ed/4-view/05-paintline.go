// paintLine is line idx as ed draws it, w wide: its number in the gutter,
// g wide, then its text from the window's first visual column, tabs run
// to their stops and the runes under the selection on a ground of their
// own.
func (v *term) paintLine(ln *outline.Node, idx, g, w int) string {
	line := []rune(ln.Text)
	lo, hi, tail := selSpan(ln)
	type cell struct {
		s   string
		sel bool
	}
	var cells []cell
	cols, vis := w-g, 0
	for i, r := range line {
		s, cw := string(r), 1
		if r == '\t' {
			cw = tabWidth - vis%tabWidth
			s = strings.Repeat(" ", cw)
		}
		if vis >= v.off+cols { // wholly right of the window
			break
		}
		if vis+cw <= v.off { // wholly left of the window
			vis += cw
			continue
		}
		if vis < v.off || vis+cw > v.off+cols { // a tab across an edge
			s = strings.Repeat(" ",
				min(vis+cw, v.off+cols)-max(vis, v.off))
		}
		cells = append(cells, cell{s, lo != -1 && i >= lo && i <= hi})
		vis += cw
	}
	if tail && vis >= v.off && vis < v.off+cols {
		cells = append(cells, cell{" ", true}) // the line break
	}
	var b strings.Builder
	b.WriteString(tty.Paint(fmt.Sprintf(" %*d  ", g-3, idx+1), tty.Quiet,
		tty.Ground, false))
	used := 0
	for j := 0; j < len(cells); {
		k := j
		var s strings.Builder
		for k < len(cells) && cells[k].sel == cells[j].sel {
			s.WriteString(cells[k].s)
			k++
		}
		bg := tty.Ground
		if cells[j].sel {
			bg = tty.Quiet
		}
		b.WriteString(tty.Paint(s.String(), tty.Fore, bg, false))
		used += tty.Width(s.String())
		j = k
	}
	b.WriteString(blank(cols - used))
	return b.String()
}
