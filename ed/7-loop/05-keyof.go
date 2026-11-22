// keyOf is a view's key as tty spells one, which ed's modes read: a rune
// when the name is one letter, else the name, and the text a paste brings.
func keyOf(k outline.Key) tty.Key {
	if r := []rune(k.Name); len(r) == 1 {
		return tty.Key{Rune: r[0], Text: k.Text}
	}
	return tty.Key{Name: k.Name, Text: k.Text}
}
