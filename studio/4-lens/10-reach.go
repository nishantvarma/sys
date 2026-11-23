// reach puts the lens on the node a key moved its cursor to, and, while
// its pane has the eye, says it to cur, as fm says its row: the shell
// beside it and e go there too. false for a node that names no place: a
// placeholder's words, a removed line, or a file gone.
func (l *lensing) reach(id string) bool {
	path, _ := outline.Address(id)
	if !filepath.IsAbs(path) {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	l.at = id
	if l.on && state.Write(l.curf, id) == nil {
		l.said = id
	}
	return true
}
