func (b *buffer) save() error {
	var sb strings.Builder
	for i, ln := range b.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(ln))
	}
	sb.WriteByte('\n')
	data := []byte(sb.String())
	if err := os.WriteFile(b.path, data, 0o644); err != nil {
		return err
	}
	b.dirty = false
	return nil
}
