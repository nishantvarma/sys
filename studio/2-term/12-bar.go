// bar is a slot's tabs as the border over its pane says them (bars), in
// tmux's styles from tty's palette, lit as sys/view lights a lens on its
// bar (`view/bar.go:/func Bar/`): each tab by its name, the one shown in
// a chip. One tab alone needs no bar.
func bar(names []string, shown int) string {
	if len(names) < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString(" ")
	for i, n := range names {
		if i == shown {
			fmt.Fprintf(&b, "#[fg=%s,bg=%s,bold] %s #[default]",
				tty.Fore, tty.Rule, n)
			continue
		}
		fmt.Fprintf(&b, "#[fg=%s] %s #[default]", tty.Soft, n)
	}
	return b.String()
}
