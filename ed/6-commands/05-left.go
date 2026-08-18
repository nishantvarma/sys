// Thin, no-arg wrappers so the keymap table can name each command as a plain
// func(*editor). They carry no logic of their own.
func (e *editor) left()     { e.step(back) }
