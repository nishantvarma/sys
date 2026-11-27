// History is capped twice: by edits, and by the rows those edits span. A
// snapshot costs one pointer per line, so the row cap is what keeps a big
// file's history from growing without bound — such a file trades undo
// depth for the ceiling instead. Under undoRows/undoDepth lines, which is
// to say almost always, only the first cap is ever felt.
const (
	undoDepth = 256
	undoRows  = 1 << 20 // ~24 MB of spine; 256 deep to 4096 lines
)
