// resolve maps a ref to a chunk key: the exact key when one matches,
// else the one key having name as a path suffix — <<paste>> finds
// buffer/paste. Two candidates are an error, not a guess.
func resolve(chunks map[string]chunk, name string) (string, bool) {
	if _, ok := chunks[name]; ok {
		return name, true
	}
	var hits []string
	for key := range chunks {
		if strings.HasSuffix(key, "/"+name) {
			hits = append(hits, key)
		}
	}
	if len(hits) > 1 {
		sort.Strings(hits)
		check(fmt.Errorf("ambiguous <<%s>>: %s",
			name, strings.Join(hits, ", ")))
	}
	if len(hits) == 0 {
		return "", false
	}
	return hits[0], true
}
