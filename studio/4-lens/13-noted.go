// noted is the text ed holds of the note at path, while it holds it
// unsaved: what the window's text file holds, for the window's unsaved
// file names that note (`ed/7-loop/14-share.go`). So the lens shows what
// is typed, not what was last saved; false when ed holds none of it.
func (l *lensing) noted(path string) ([]byte, bool) {
	if state.Read(l.unsavedf) != path {
		return nil, false
	}
	b, err := os.ReadFile(l.textf)
	return b, err == nil
}
