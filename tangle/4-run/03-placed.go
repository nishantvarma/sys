placed := map[string]bool{}
note := func(src string) {
	for _, m := range ref.FindAllStringSubmatch(src, -1) {
		if esc, n := m[1], m[2]; esc == "" && n != "others" {
			if key, ok := resolve(chunks, n); ok {
				placed[key] = true
			}
		}
	}
}
note(string(root))
for _, c := range chunks {
	note(c.body)
}
