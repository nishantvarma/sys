// share puts the text ed holds in the window's text file while it is
// unsaved, so that the lens beside ed shows what is typed and not what was
// last saved (`studio/4-lens/13-noted.go`). A sum keeps a move, which
// changes nothing, from waking the lens; mark takes the file away with
// the unsaved one. Outside a frame there is none to keep.
func (e *editor) share() {
	if e.text == "" || !e.b.dirty {
		return
	}
	b := e.b.bytes()
	sum := sha256.Sum256(b)
	if sum == e.shared {
		return
	}
	if err := state.Put(e.text, b); err != nil {
		e.msg = err.Error()
		return
	}
	e.shared = sum
}
