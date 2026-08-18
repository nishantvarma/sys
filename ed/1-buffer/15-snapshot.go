// History is capped twice: by edits, and by the rows those edits span. A
// snapshot costs one pointer per line, so the row cap is what keeps a big
// file's history from growing without bound — such a file trades undo
// depth for the ceiling instead. Under undoRows/undoDepth lines, which is
// to say almost always, only the first cap is ever felt.
const (
	undoDepth = 256
	undoRows  = 1 << 20 // ~24 MB of spine; 256 deep to 4096 lines
)

// snapshot records the state an edit is about to leave behind. c is the
// caret from before the edit, so undo returns the cursor to what the edit
// was aimed at.
//
// Only the spine is copied. Lines the edit never touches are shared with
// history rather than duplicated, so an edit costs its own size and not
// the file's. That is the rule the rest of the buffer keeps in exchange: a
// line's runes are never written in place, only replaced — insert,
// backspace and replaceChar each clone the one line they change.
func (b *buffer) snapshot(c sel) {
	b.undo = append(b.undo, snap{slices.Clone(b.lines), c})
	rows := 0
	for _, s := range b.undo {
		rows += len(s.lines)
	}
	for len(b.undo) > 1 &&
		(len(b.undo) > undoDepth || rows > undoRows) {
		rows -= len(b.undo[0].lines)
		// Reslicing alone would leave the evicted spine reachable
		// from the array it still sits in, and so uncollectable.
		b.undo[0] = snap{}
		b.undo = b.undo[1:]
	}
	b.redo = nil // a fresh edit branches history; the redo stack is stale
}
