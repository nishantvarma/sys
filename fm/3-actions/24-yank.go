func (m *fm) yank(content bool) {
	c := m.cur()
	if c == "" {
		return
	}
	var data []byte
	if content && isFile(c) {
		data, _ = os.ReadFile(c)
	} else {
		abs, _ := filepath.Abs(c)
		data = []byte(abs)
	}
	m.catch(m.ui.Copy(string(data)))
}
