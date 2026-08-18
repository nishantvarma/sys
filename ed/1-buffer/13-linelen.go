// lineLen is len(line(i)), asked separately because it is asked constantly
// — every clamp, every line end, every step off the end of a row — and a
// count is the one thing a line can answer without being materialised.
func (b *buffer) lineLen(i int) int { return len(b.lines[i]) }
