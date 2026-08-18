func classAt(b *buffer, p pos) int {
	r, ok := runeAt(b, p)
	switch {
	case !ok, r == '\n', unicode.IsSpace(r):
		return clSpace
	case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_':
		return clWord
	default:
		return clPunct
	}
}
