// detach launches a terminal program in the background, off our tty.
func (m *fm) detach(cmd ...string) {
	c := exec.Command("spawn", append([]string{termProg}, cmd...)...)
	c.Dir = m.ctx()
	if c.Start() == nil {
		go c.Wait() // reap it; don't leak a zombie per detach
	}
}
