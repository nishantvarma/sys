func commonPrefix(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	p := ss[0]
	for _, s := range ss[1:] {
		i := 0
		for i < len(p) && i < len(s) && p[i] == s[i] {
			i++
		}
		p = p[:i]
	}
	return p
}
