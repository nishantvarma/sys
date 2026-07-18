func render(v view) string {
	var b strings.Builder
	b.WriteString(tty.Home)
	tty.Line(&b, header(v))
	h := v.h - 2
	if len(v.files) > h {
		h--
	}
	if h < 0 {
		h = 0
	}
	off := tty.ScrollOffset(v.idx, len(v.files), h)
	vis := v.files[off:min(off+h, len(v.files))]
	for i, f := range vis {
		tty.Line(&b, row(off+i, f, v))
	}
	extra := h - len(vis)
	if len(v.files) > h {
		tty.Line(&b, tty.Dim(fmt.Sprintf(" +%d", len(v.files)-h)))
		extra--
	}
	for i := 0; i < extra; i++ {
		tty.Line(&b, "")
	}
	msg := v.msg
	if r := []rune(msg); len(r) > v.w {
		// clamp to width so a long flash can't scroll the frame
		msg = string(r[:v.w])
	}
	b.WriteString(tty.Status(v.h, msg, false))
	return b.String()
}
