// facts is what fm says beside a thing, as the samples said it: a file's
// size and how long ago it changed, a link's target. A folder has none,
// for its count would cost a read each.
func facts(f entry) []string {
	switch {
	case f.link:
		if t, err := os.Readlink(f.path); err == nil {
			return []string{"→ " + t}
		}
		return nil
	case f.dir:
		return nil
	}
	return []string{fmt.Sprintf("%-5s %s", size(f.size),
		age(time.Since(f.mod)))}
}
