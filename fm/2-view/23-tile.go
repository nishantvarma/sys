// tile is row r as tile i: its number, then its label in its kind's look,
// cut or padded to cw so that the next tile lines up.
func tile(i int, r row, cw int) string {
	color, _ := look(r.n.Kind)
	name := tty.Clip(label(r.n), cw)
	pad := strings.Repeat(" ", max(0, cw-tty.Width(name)))
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	return " " + num + " " + hl(r.n) + color(name) + pad + " "
}
