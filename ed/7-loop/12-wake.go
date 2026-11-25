// wake has ed hear cur change while it waits on a key, when it chases cur
// on the terminal: a watcher on cur's folder wakes the read as a resize
// does (tty's Wake), and the loop chases before it draws again. Down a
// pipe the host wakes it with a size, which the stream reads as a resize,
// as gio's window does (`gio/host.go:/a size wakes a program/`). A folder
// it cannot watch leaves ed chasing only as keys come.
func (e *editor) wake() {
	if e.curf == "" || e.t == nil {
		return
	}
	dir := filepath.Dir(e.curf)
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	w, err := watch.New(dir)
	if err != nil {
		return
	}
	go func() {
		for range w.On {
			e.t.Wake()
		}
	}()
}
