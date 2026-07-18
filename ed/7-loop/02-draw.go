func (e *editor) draw() {
	w, h := e.t.Size()
	e.scroll(h-1, w)
	a, c := e.rng()
	mode := e.mode.name()
	if e.ext {
		mode = "extend"
	}
	e.t.Write(render(frame{
		lines: e.b.lines,
		a:     a,
		c:     c,
		head:  e.head(),
		sel:   a != c,
		top:   e.top,
		off:   e.off,
		w:     w,
		h:     h,
		mode:  mode,
		path:  e.b.path,
		dirty: e.b.dirty,
		msg:   e.msg,
	}))
	e.t.Flush()
}
