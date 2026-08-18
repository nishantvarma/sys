// save writes the lines out, a break after each, over the file ed last
// read or wrote, and over nothing else. Text another writer put there
// stays until R has read it; u then brings the buffer back over it, and
// save writes, the file being one ed has read. No file is nothing to
// lose, so a new one is written.
func (b *buffer) save() error {
	on, err := os.ReadFile(b.path)
	if err == nil && sha256.Sum256(on) != b.disk {
		return errors.New("changed on disk — " +
			"R reads it, u brings yours back")
	}
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
	b.disk = sha256.Sum256(data)
	b.dirty = false
	return nil
}
