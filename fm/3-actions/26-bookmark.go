// bookmark says the row out loud: fm writes cur but never the log, so only
// c, the one appender, makes a browsed row findable by m. A content hit
// still carries its line, the way ed's save does.
func (m *fm) bookmark() {
	c := m.cur()
	if c == "" {
		return
	}
	if n := m.hitLine(c); n > 0 {
		c += ":" + strconv.Itoa(n)
	}
	cmd := exec.Command("c", c)
	cmd.Dir = m.ctx()
	if cmd.Start() == nil {
		go cmd.Wait() // reap it; don't leak a zombie per press
		m.flash("bookmarked " + filepath.Base(c))
	}
}
