// face is the frame's face, the slot that runs fm: its end ends the frame
// (closeVerb), and fm's q asks first while ed holds work unsaved; nil
// when no slot runs fm.
func (f *frame) face() *slot {
	for _, s := range leaves(f.slots) {
		if w := strings.Fields(s.cmd); len(w) > 0 && w[0] == "fm" {
			return s
		}
	}
	return nil
}
