// bar is the top line, w wide: each lens in a chip, the one shown lit,
// and at the right tab, the key that turns to the next; and where each
// chip stands, the lens a press there turns to, tab's the next. A chip
// that does not fit is left off.
func (v *term) bar(w int) (string, []chip) {
	var b strings.Builder
	var at []chip
	x := 0
	put := func(s, fg, bg string, bold bool) {
		b.WriteString(tty.Paint(s, fg, bg, bold))
		x += tty.Width(s)
	}
	put(" ", "", tty.Ground, false)
	for i, name := range lenses {
		c := " " + name + " "
		if x+tty.Width(c) > w {
			break
		}
		fg, bg := tty.Soft, tty.Ground
		if i == v.lens {
			fg, bg = tty.Fore, tty.Rule
		}
		at = append(at, chip{x, x + tty.Width(c), i})
		put(c, fg, bg, i == v.lens)
	}
	key, says := " tab ", " next lens"
	if room := w - x - tty.Width(key+says) - 1; room >= 0 {
		put(strings.Repeat(" ", room), "", tty.Ground, false)
		next := (v.lens + 1) % len(lenses)
		at = append(at, chip{x, x + tty.Width(key), next})
		put(key, tty.Fore, tty.Rule, false)
		put(says, tty.Soft, tty.Ground, false)
	}
	put(strings.Repeat(" ", max(0, w-x)), "", tty.Ground, false)
	return b.String(), at
}

// chip is where a press on the bar falls, columns from and up to to,
// counted from 0, and the lens it turns to.
type chip struct {
	from, to, lens int
}

// tops is the lines over the text, h high: the bar and a rule under it,
// as fm's view keeps its head and a rule (sys/view's chrome); the bar
// alone in a pane under twelve lines high, as the view keeps its head
// alone there; and neither in a pane too short to spare a line.
func tops(h int) int {
	switch {
	case h < 3:
		return 0
	case h < 12:
		return 1
	}
	return 2
}

// ruled is the rule under the bar, w wide, as sys/view draws its own.
func ruled(w int) string {
	return tty.Paint(strings.Repeat("─", max(0, w)), tty.Rule, tty.Ground,
		false)
}

// onBar is whether k is a press on the bar, and for a left one the lens
// its chip turns to, -1 between chips or for any other press there: the
// bar is no line and no item. A button let go there goes on as any does.
func (v *term) onBar(k tty.Key) (int, bool) {
	w, h := v.t.Size()
	if k.Name != tty.KMouse || k.Y != 1 || tops(h) == 0 {
		return 0, false
	}
	if k.Button != tty.MLeft {
		return -1, true
	}
	_, at := v.bar(w)
	for _, c := range at {
		if c.from <= k.X-1 && k.X-1 < c.to {
			return c.lens, true
		}
	}
	return -1, true
}
