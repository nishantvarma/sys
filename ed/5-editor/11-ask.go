// ask shows msg and reads one more key for the verb under way — r's rune,
// f's target. The key joins typed, so a replay of r finds its rune queued
// right behind it instead of waiting for one typed live. A prompt's line
// is not kept: / and : change nothing, so . never needs it. A resize, or
// a waking as cur changes, is no answer: ed draws again, and asks on.
func (e *editor) ask(msg string) (tty.Key, bool) {
	e.msg = msg
	for {
		e.draw()
		ev, ok := e.ui.Key()
		k := keyOf(ev)
		if ok && k.Name == tty.KResize {
			continue
		}
		e.msg = ""
		e.typed = append(e.typed, k)
		return k, ok
	}
}
