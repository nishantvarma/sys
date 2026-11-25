// modeWord is the key a press on the status line's first word is: the
// mode it names, as a button, esc out of insert and i into it, so the
// mouse goes between them as the keys do; "" for a press anywhere else.
// x and y count from 1, as the terminal counts them.
func (v *term) modeWord(x, y int) string {
	_, h := v.t.Size()
	if y != h || x > len(v.mode)+1 {
		return ""
	}
	if v.mode == "insert" {
		return tty.KEsc
	}
	return "i"
}
