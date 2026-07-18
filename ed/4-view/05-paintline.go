// paintLine expands line idx and clips it to the window [f.off, f.off+f.w)
// of visual columns, highlighting the runes under the selection.
func paintLine(f frame, idx int) string {
	line := f.lines[idx]
	lo, hi, tail := selSpan(f, idx)
	var sb strings.Builder
	vis, inSel := 0, false
	for i, r := range line {
		cell, w := " ", 1
		if r == '\t' {
			w = tabWidth - vis%tabWidth
			cell = strings.Repeat(" ", w)
		} else {
			cell = string(r)
		}
		if vis >= f.off+f.w { // wholly right of the window
			break
		}
		if vis+w <= f.off { // wholly left of the window
			vis += w
			continue
		}
		if vis < f.off || vis+w > f.off+f.w { // a tab across an edge
			cell = strings.Repeat(" ",
				min(vis+w, f.off+f.w)-max(vis, f.off))
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
		vis += w
	}
	if inSel {
		sb.WriteString(tty.Sgr0)
	}
	if tail && vis >= f.off && vis < f.off+f.w {
		sb.WriteString(
			selBg + " " + tty.Sgr0,
		) // the selected line break
	}
	return sb.String()
}
