// where is the buffer's path as an address starts: absolute, as fm's and
// cur's are, whatever ed was given; empty for a buffer with no file.
func (e *editor) where() string {
	if e.b.path == "" {
		return ""
	}
	if p, err := filepath.Abs(e.b.path); err == nil {
		return p
	}
	return e.b.path
}
