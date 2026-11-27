// wake has ed hear cur change while it waits on a key, when it chases cur
// on the terminal: cur's folder wakes the read as a resize does
// (watch.Wake), and the loop chases before it draws again. Down a pipe
// the host wakes it with a size, which the stream reads as a resize, as
// gio's window does (`gio/host.go:/a size wakes a program/`).
func (e *editor) wake() {
	if e.curf != "" && e.t != nil {
		watch.Wake(filepath.Dir(e.curf), e.t.Wake)
	}
}
