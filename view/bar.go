package view

import "sys/tty"

// Bar is the line a program with lenses shows over what it draws
// (`design/lens`), w wide: each lens named in a chip, the one shown lit,
// and at the right tab, the key that turns to the next; and where each
// chip stands, the lens a press there names, tab's the next. A chip that
// does not fit is left off. ed's terminal view and gio's window draw it
// alike, from the names a program puts on its tree's tags.
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
	key, says := " tab ", " next lens"
	if room := w - k.w - tty.Width(key+says) - 1; room >= 0 &&
		len(names) > 0 {
		k.pad(k.w+room, "")
		next := names[(on+1)%len(names)]
		at = append(at, Chip{k.w, k.w + tty.Width(key), next})
		k.add(key, tty.Fore, tty.Rule, false)
		k.add(says, tty.Soft, "", false)
	}
	return k.clip(w).pad(w, "").runs, at
}

// Chip is where a press on the bar falls, from column From up to To,
// counted from 0, and the lens it names.
type Chip struct {
	From, To int
	Lens     string
}

// Shown is which of the lenses named shows: the one the program's state
// says, as ed says any but its first, else the first.
func Shown(names, state []string) int {
	for i, name := range names {
		for _, s := range state {
			if s == name {
				return i
			}
		}
	}
	return 0
}

// Tops is the lines the bar and its rule take over a program's text, h
// high, as the view keeps its head and a rule over its rows (chrome): the
// two, the bar alone in a pane under twelve lines high, as the view keeps
// its head alone there, and neither in one too short to spare a line.
func Tops(h int) int {
	switch top, _ := chrome(h); {
	case h < 3:
		return 0
	case top == 1:
		return 1
	}
	return 2
}

// Rule is a line across, w wide, in the rule's colour, as the view draws
// under its head, and a bar's host under the bar.
func Rule(w int) []Run {
	return rule(w).ink.runs
}
