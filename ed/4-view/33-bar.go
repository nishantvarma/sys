// bar is the top line, w wide, as sys/view draws it (view.Bar): each lens
// ed names on the tree it drew last, the one its state says lit, and tab
// at the right; and where each chip stands, the lens a press there names.
func (v *term) bar(w int) (string, []view.Chip) {
	var names []string
	if v.last != nil {
		names = v.last.Tags
	}
	runs, at := view.Bar(names, view.Shown(names, v.state), w)
	return view.Paint(runs), at
}

// tops is the lines over the text, h high: the bar and the rule under
// it, as many as sys/view gives them (view.Tops).
func tops(h int) int {
	return view.Tops(h)
}

// ruled is the rule under the bar, w wide, as sys/view draws its own.
func ruled(w int) string {
	return view.Paint(view.Rule(w))
}

// onBar is whether k is a press on the bar, and for a left one the lens
// its chip names, none between chips or for any other press there: the
// bar is no line and no item. A button let go there goes on as any does.
func (v *term) onBar(k tty.Key) (string, bool) {
	w, h := v.t.Size()
	if k.Name != tty.KMouse || k.Y != 1 || tops(h) == 0 {
		return "", false
	}
	if k.Button != tty.MLeft {
		return "", true
	}
	_, at := v.bar(w)
	for _, c := range at {
		if c.From <= k.X-1 && k.X-1 < c.To {
			return c.Lens, true
		}
	}
	return "", true
}
