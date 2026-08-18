func (m *fm) draw() {
	w, h := m.t.Size()
	m.t.Write(render(view{
		cwd:     m.cwd,
		files:   m.files,
		idx:     m.idx,
		sel:     m.sel,
		clip:    len(m.clip),
		cutting: m.cutting,
		alt:     m.alt,
		msg:     m.msg,
		w:       w,
		h:       h,
	}))
	m.t.Flush()
}
