// written is the line over ed's text, w wide: that it shows the thing as
// written, and which lines the cursor's item spans, from home, as the
// samples said a source: as written · notes/plan:3..5.
func written(path string, a, z, w int) string {
	if userHome != "" {
		path = strings.TrimPrefix(path, userHome+"/")
	}
	span := fmt.Sprintf("%s:%d", path, a+1)
	if z > a {
		span += fmt.Sprintf("..%d", z+1)
	}
	used := tty.Width(" as written · " + span)
	return tty.Paint(" as written", tty.Soft, tty.Ground, false) +
		tty.Paint(" · ", tty.Quiet, tty.Ground, false) +
		tty.Paint(span, tty.Soft, tty.Ground, false) +
		blank(max(0, w-used))
}
