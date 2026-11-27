package view

import "sys/tty"

// Bar is a slot's tab line, as gio's window draws one over a slot of tabs
// and tmux's border holds one (`studio/2-term/12-bar/01-bar.go`), w wide:
// each tab named in a chip, the one at on lit; and where each chip
// stands, the name a press there says. A chip that does not fit is left
// off.
func Bar(names []string, on, w int) ([]Run, []Chip) {
	k := (&ink{}).add(" ", "", "", false)
	var at []Chip
	for i, name := range names {
		c := " " + printable(name) + " "
		if k.w+tty.Width(c) > w {
			break
		}
		fg, bg := tty.Soft, ""
		if i == on {
			fg, bg = tty.Fore, tty.Rule
		}
		at = append(at, Chip{k.w, k.w + tty.Width(c), name})
		k.add(c, fg, bg, i == on)
	}
	return k.clip(w).pad(w, "").runs, at
}

// Chip is where a press on the bar falls, from column From up to To,
// counted from 0, and the name a press there says.
type Chip struct {
	From, To int
	Name     string
}
