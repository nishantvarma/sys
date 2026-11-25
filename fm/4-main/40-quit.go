// quit ends fm, and in a frame's slot the frame with it, for fm is its
// face (`studio/5-verbs/10-closeverb.go:/closeVerb/`): but while ed names
// a file it holds unsaved (`ed/7-loop/13-mark.go`), the first q only says
// so, and the next lets the work go, taking the name away; a frame whose
// face ends with the name still there, as after a crash, stays for ed to
// save. Outside a frame there is no file to read, and it reads as nothing.
func (m *fm) quit() {
	if f := state.Read(m.unsaved); f != "" {
		if !m.armed {
			m.armed = true
			m.flash("unsaved in ed: " + filepath.Base(f) +
				" — q again to discard")
			return
		}
		os.Remove(m.unsaved) // the work let go
	}
	m.done = true
}
