func (m *fm) zoxide() {
	q, ok := m.prompt("z: ", nil, "")
	if !ok || q == "" {
		return
	}
	out, err := exec.Command(
		"zoxide",
		append([]string{"query"}, strings.Fields(q)...)...,
	).Output()
	if err != nil {
		return
	}
	if r := strings.TrimSpace(string(out)); r != "" {
		m.cd(r)
	}
}
