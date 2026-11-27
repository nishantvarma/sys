// bytes is the lines as save writes them, a break after each: the file
// as it would be, which the lens reads while it is unsaved (share).
func (b *buffer) bytes() []byte {
	var sb strings.Builder
	for i, ln := range b.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(ln))
	}
	sb.WriteByte('\n')
	return []byte(sb.String())
}
