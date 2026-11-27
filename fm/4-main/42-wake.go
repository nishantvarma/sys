// wake has fm hear cur change while it waits on a key, when it follows cur
// on the terminal: cur's folder wakes the read as a resize does
// (watch.Wake), and the loop follows before it draws again, as ed's wake
// does. Down a pipe the host wakes it with a size.
func (m *fm) wake() {
	if m.follows && m.t != nil {
		watch.Wake(filepath.Dir(m.this), m.t.Wake)
	}
}
