func render(v view) string {
	var b strings.Builder
	b.WriteString(tty.Home)
	tty.Line(&b, tty.Clip(header(v), v.w))
	h := v.h - 2
	if len(v.files) > h {
		h--
	}
	if h < 0 {
		h = 0
	}
	off := max(0, min(v.idx-h/2, len(v.files)-h)) // cursor row centred
	vis := v.files[off:min(off+h, len(v.files))]
	for i, f := range vis {
		tty.Line(&b, tty.Clip(row(off+i, f, v), v.w))
	}
	extra := h - len(vis)
	if len(v.files) > h {
		tty.Line(&b, tty.Dim(fmt.Sprintf(" +%d", len(v.files)-h)))
		extra--
	}
	for i := 0; i < extra; i++ {
		tty.Line(&b, "")
	}
	b.WriteString(tty.Status(v.h, tty.Clip(v.msg, v.w), false))
	return b.String()
}
