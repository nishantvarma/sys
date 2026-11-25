// quit ends fm, and in a frame's slot the frame with it, for fm is its
// face (`studio/5-verbs/10-closeverb.go:/closeVerb/`): but while ed names
// a file it holds unsaved (`ed/7-loop/13-mark.go`), the first q only says
// so, and the next ends it all, the work discarded. Outside a frame there
// is no file to read, and it reads as nothing.
func (m *fm) quit() {
	if f := state.Read(m.unsaved); f != "" && !m.armed {
		m.armed = true
		m.flash("unsaved in ed: " + filepath.Base(f) +
			" — q again to discard")
		return
	}
	m.done = true
}
