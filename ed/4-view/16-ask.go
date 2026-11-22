// Ask reads a line at the status bar. Completion stays tty's seam, which
// ed leaves empty.
func (v *term) Ask(prompt, def string) (string, bool) {
	return tty.ReadLine(v.t, prompt, nil, def)
}
