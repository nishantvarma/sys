// take is the lens's next event. On the terminal it is the view's next
// key, and a change wakes the read as a resize does (tty's Wake), so the
// view stays in one goroutine. Down a pipe the keys come from a goroutine,
// and a change is a resize too.
func take(ui outline.View, keys chan keyed, wt *watch.Watcher) (
	outline.Key, bool) {
	if keys == nil {
		return ui.Key()
	}
	select {
	case k := <-keys:
		return k.ev, k.ok
	case <-wt.On:
		return outline.Key{Name: tty.KResize}, true
	}
}
