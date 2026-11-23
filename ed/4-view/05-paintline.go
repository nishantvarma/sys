// paintLine is line idx as ed draws it, w wide: its number in the gutter,
// then its text from the window's first visual column, tabs run to their
// stops and the runes under the selection on a ground of their own. A
// line of the cursor's item is lit across, its number brighter, and on
// the cursor's own line the label, up to its note, is bold.
func (v *term) paintLine(ln *outline.Node, idx int, lit, cur bool,
	w int) string {
	bg, num := tty.Ground, tty.Quiet
	if lit {
		bg, num = tty.Bar, tty.Soft
	}
	line := []rune(ln.Text)
	lo, hi, tail := selSpan(ln)
	la, lb := -1, -1
	if cur {
		la, lb = labelOf(line)
	}
	type cell struct {
		s         string
		sel, bold bool
	}
	var cells []cell
	cols, vis := w-gutter, 0
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
		sel := lo != -1 && i >= lo && i <= hi
		cells = append(cells, cell{s, sel, i >= la && i < lb})
		vis += cw
	}
	if tail && vis >= v.off && vis < v.off+cols {
		cells = append(cells, cell{" ", true, false}) // the line break
	}
	var b strings.Builder
	b.WriteString(tty.Paint(fmt.Sprintf(" %3d  ", idx+1), num, bg,
		false))
	used := 0
	for j := 0; j < len(cells); {
		k := j
		var s strings.Builder
		for k < len(cells) && cells[k].sel == cells[j].sel &&
			cells[k].bold == cells[j].bold {
			s.WriteString(cells[k].s)
			k++
		}
		g := bg
		if cells[j].sel {
			g = tty.Quiet
		}
		bold := cells[j].bold
		b.WriteString(tty.Paint(s.String(), tty.Fore, g, bold))
		used += tty.Width(s.String())
		j = k
	}
	pad := strings.Repeat(" ", max(0, cols-used))
	b.WriteString(tty.Paint(pad, "", bg, false))
	return b.String()
}
