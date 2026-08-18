// ask shows msg and reads one more key for the verb under way — r's rune,
// f's target. The key joins typed, so a replay of r finds its rune queued
// right behind it instead of waiting for one typed live. A prompt's line
// is not kept: / and : change nothing, so . never needs it.
func (e *editor) ask(msg string) (tty.Key, bool) {
	e.msg = msg
	e.draw()
	k, ok := e.t.ReadKey()
	e.msg = ""
	e.typed = append(e.typed, k)
	return k, ok
}
