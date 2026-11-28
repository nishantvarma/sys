// watched has fm hear its folder change, so that it reads it again then,
// or after a verb, and not at every key (loop): its watch follows the
// folder fm stands in (ls), and on the terminal wakes the read as a
// resize does; down a pipe the next key or wake finds the folder stale.
// With no watch, fm reads it at every key, as it did.
func (m *fm) watched() {
	w, err := watch.New()
	if err != nil {
		return
	}
	m.watcher = w
	m.stale.Store(true)
	go func() {
		for range w.On {
			m.stale.Store(true)
			if m.t != nil {
				m.t.Wake()
			}
		}
	}()
}
