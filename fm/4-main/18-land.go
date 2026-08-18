// land puts the cursor on name, else where a first visit starts: after ..
func (m *fm) land(name string) {
	m.idx = m.at(name)
	if m.idx < 0 {
		m.idx = min(m.at("..")+1, max(0, len(m.files)-1))
	}
}
