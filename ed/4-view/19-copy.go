// Copy puts text on X's clipboard.
func (v *term) Copy(text string) error {
	return tty.Put(tty.Clipboard, text)
}
