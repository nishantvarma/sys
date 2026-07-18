// pasteDir inserts reg after the cursor, or before it when before is
// set (P). An active selection replaces it instead, direction no
// longer meaning anything; the cut register is discarded so reg
// survives for the paste.
func (e *editor) pasteDir(before bool) {
	if len(e.reg) == 0 {
		return
	}
	e.b.snapshot()
	if a, c := e.rng(); a != c {
		e.b.cut(a, c)
		e.place(back(e.b, e.b.paste(e.b.clamp(a, true), e.reg)))
		return
	}
	h := e.head()
	if r := e.reg; r[len(r)-1] == '\n' { // line-wise: no mid-line split
		at := pos{h.line, len(e.b.lines[h.line])}
		nl := append([]rune{'\n'}, r[:len(r)-1]...)
		if before {
			at, nl = pos{h.line, 0}, r
		}
		e.place(back(e.b, e.b.paste(at, nl)))
		return
	}
	at := e.b.clamp(pos{h.line, h.col + 1}, true)
	if before {
		at = h
	}
	e.place(back(e.b, e.b.paste(at, e.reg)))
}
