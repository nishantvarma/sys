// blank is w blanks on the palette's ground.
func blank(w int) string {
	return tty.Paint(strings.Repeat(" ", max(0, w)), "", tty.Ground, false)
}
