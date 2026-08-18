func (m *fm) targets() []string {
	if len(m.sel) > 0 {
		out := make([]string, 0, len(m.sel))
		for p := range m.sel {
			out = append(out, p)
		}
		slices.Sort(out) // map order is random; keep rm/paste stable
		return out
	}
	if c := m.cur(); c != "" {
		return []string{c}
	}
	return nil
}
