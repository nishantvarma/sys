func row(i int, f entry, v view) string {
	color, suf := style(f)
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	name := color(f.name + suf)
	return fmt.Sprintf(" %s  %s%s", num, hl(i, f, v), name)
}
