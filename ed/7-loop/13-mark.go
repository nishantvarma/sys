// mark keeps the window's unsaved file true while ed runs in a frame's
// slot: it names ed's file while the buffer is unsaved, and is gone once
// it is saved or ed ends, so that fm's q asks before it ends the frame
// (`fm/4-main/40-quit.go`); the text file, which share keeps, goes with
// it, and the lens reads the file again. Outside a frame there is none
// to keep.
func (e *editor) mark(unsaved bool) {
	if e.unsaved == "" || unsaved == e.marked {
		return
	}
	e.marked = unsaved
	if !unsaved {
		os.Remove(e.unsaved)
		os.Remove(e.text)
		e.shared = [32]byte{}
		return
	}
	name := e.where()
	if name == "" {
		name = "a buffer with no file"
	}
	if err := state.Write(e.unsaved, name); err != nil {
		e.msg = err.Error()
	}
}
