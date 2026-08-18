func (t *tangler) others(dir, self string, stack []string) string {
	var names []string
	for name, c := range t.chunks {
		if name == self || t.placed[name] || t.swept[name] {
			continue
		}
		if dir == "." || strings.HasPrefix(c.path, dir+"/") {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return t.chunks[names[i]].path < t.chunks[names[j]].path
	})

	var parts []string
	for _, name := range names {
		t.swept[name] = true
		c := t.chunks[name]
		parts = append(parts, strings.TrimRight(t.expand(
			c.body, filepath.Dir(c.path), name,
			append(stack, name)), "\n"))
	}
	return strings.Join(parts, "\n\n")
}
