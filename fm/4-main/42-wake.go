// wake has fm hear cur change while it waits on a key, when it follows cur
// on the terminal: a watcher on cur's folder wakes the read as a resize
// does (tty's Wake), and the loop follows before it draws again, as ed's
// wake does. Down a pipe the host wakes it with a size. A folder it
// cannot watch leaves fm following only as keys come.
func (m *fm) wake() {
	if !m.follows || m.t == nil {
		return
	}
	dir := filepath.Dir(m.this)
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	w, err := watch.New(dir)
	if err != nil {
		return
	}
	go func() {
		for range w.On {
			m.t.Wake()
		}
	}()
}
