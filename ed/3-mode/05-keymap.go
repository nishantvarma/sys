func keymap(bs []binding) map[string]func(*editor) {
	m := make(map[string]func(*editor), len(bs))
	for _, b := range bs {
		m[b.key] = b.act
	}
	return m
}
