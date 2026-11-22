// Copy mirrors text to the system clipboard via OSC 52; terminals that
// don't support it just ignore the escape.
func (v *term) Copy(text string) {
	b64 := base64.StdEncoding.EncodeToString([]byte(text))
	v.t.Write("\x1b]52;c;" + b64 + "\x07")
}
