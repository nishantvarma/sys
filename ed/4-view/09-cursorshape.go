// cursorShape: a beam where the caret sits between runes (insert), a block
// where it sits on one (normal).
func cursorShape(name string) string {
	if name == "insert" {
		return tty.CurBeam
	}
	return tty.CurBlock
}
