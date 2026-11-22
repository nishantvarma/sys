// paintLine expands line ln and clips it to the window [v.off, v.off+w) of
// visual columns, highlighting the runes under the selection.
func (v *term) paintLine(ln *outline.Node, w int) string {
	line := []rune(ln.Text)
	lo, hi, tail := selSpan(ln)
	var sb strings.Builder
	vis, inSel := 0, false
	for i, r := range line {
		cell, cw := " ", 1
		if r == '\t' {
			cw = tabWidth - vis%tabWidth
			cell = strings.Repeat(" ", cw)
		} else {
			cell = string(r)
		}
		if vis >= v.off+w { // wholly right of the window
			break
		}
		if vis+cw <= v.off { // wholly left of the window
			vis += cw
			continue
		}
		if vis < v.off || vis+cw > v.off+w { // a tab across an edge
			cell = strings.Repeat(" ",
				min(vis+cw, v.off+w)-max(vis, v.off))
		}
		if sel := lo != -1 && i >= lo && i <= hi; sel != inSel {
			if sel {
				sb.WriteString(selBg)
			} else {
				sb.WriteString(tty.Sgr0)
			}
			inSel = sel
		}
		sb.WriteString(cell)
		vis += cw
	}
	if inSel {
		sb.WriteString(tty.Sgr0)
	}
	if tail && vis >= v.off && vis < v.off+w {
		sb.WriteString(
			selBg + " " + tty.Sgr0,
		) // the selected line break
	}
	return sb.String()
}
