// searchSel sets the pattern from the selected text — one trailing
// newline stripped — and searches for its next occurrence.
func (e *editor) searchSel() {
	pat := e.b.text(e.rng())
	if n := len(pat); n > 0 && pat[n-1] == '\n' {
		pat = pat[:n-1]
	}
	if len(pat) == 0 {
		return
	}
	e.pat = string(pat)
	e.searchNext()
}
