// hitLine is the line a pending content hit brought, when it belongs to
// p — the one fact edit and bookmark both read.
func (m *fm) hitLine(p string) int {
	if m.hit == p {
		return m.line
	}
	return 0
}
