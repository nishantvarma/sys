func (t *tangler) expand(src, dir, self string, stack []string) string {
	return ref.ReplaceAllStringFunc(src, func(m string) string {
		sub := ref.FindStringSubmatch(m)
		esc, name := sub[1], sub[2]
		if esc != "" {
			return "<<" + name + ">>"
		}
		if name == "others" {
			return t.others(dir, self, stack)
		}
		// a name that names no chunk would be left in the output, to
		// surface three steps on as a compile error: say it here. what
		// is no name, <<%s>> in a format, stays as it is
		key, ok := resolve(t.chunks, name)
		if !ok && named.MatchString(name) {
			in := "@main.go"
			if c, ok := t.chunks[self]; ok {
				in = c.path
			}
			check(fmt.Errorf("%s: <<%s>> names no chunk",
				in, name))
		}
		if !ok {
			return m
		}
		c := t.chunks[key]
		for _, s := range stack {
			if s == key {
				check(fmt.Errorf(
					"cycle: <<%s>> "+
						"ultimately refers to itself",
					key))
			}
		}
		t.expanded[key] = true
		return strings.TrimRight(t.expand(
			c.body, filepath.Dir(c.path), key,
			append(stack, key)), "\n")
	})
}
