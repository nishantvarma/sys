// yank copies the selection into reg, and mirrors it to the system
// clipboard via OSC 52 — terminals that don't support it just ignore
// the escape.
func (e *editor) yank() {
	a, c := e.rng()
	e.reg = e.b.text(a, c)
	b64 := base64.StdEncoding.EncodeToString([]byte(string(e.reg)))
	e.t.Write("\x1b]52;c;" + b64 + "\x07")
}
