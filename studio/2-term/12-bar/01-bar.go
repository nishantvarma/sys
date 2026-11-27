// bar is a slot's tabs as the border over its pane says them (bars), in
// tmux's styles from tty's palette, lit as sys/view lights a lens on its
// bar (`view/bar.go:/func Bar/`): each tab by its name, the one shown in
// a chip. Each name is a button, a control range by its place, and the
// panel's bar ends in two more, + and, with a tab to spare, × (tabButtons).
// A slot of one tab needs no bar, but the panel does, for its +.
func bar(names []string, shown int, panel bool) string {
	if len(names) < 2 && !panel {
		return ""
	}
	var b strings.Builder
	b.WriteString(" ")
	button := func(n int, chip string) {
		fmt.Fprintf(&b, "#[range=control|%d]%s#[norange]", n, chip)
	}
	soft := func(s string) string {
		return fmt.Sprintf("#[fg=%s] %s #[default]", tty.Soft, s)
	}
	for i, n := range names {
		chip := soft(n)
		if i == shown {
			chip = fmt.Sprintf("#[fg=%s,bg=%s,bold] %s #[default]",
				tty.Fore, tty.Rule, n)
		}
		if i >= tabButtons {
			b.WriteString(chip)
			continue
		}
		button(i, chip)
	}
	if panel {
		button(newButton, soft("+"))
		if len(names) > 1 {
			button(closeButton, soft("×"))
		}
	}
	return b.String()
}
