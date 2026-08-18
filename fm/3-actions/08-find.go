func (m *fm) find(d int) {
	if m.pat == "" || len(m.files) == 0 {
		return
	}
	n := len(m.files)
	for i := 1; i <= n; i++ {
		idx := ((m.idx+d*i)%n + n) % n
		if strings.Contains(
			strings.ToLower(m.files[idx].name),
			m.pat,
		) {
			m.idx = idx
			return
		}
	}
}
