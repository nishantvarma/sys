// binding maps a key to an action. A "" desc hides it from the help screen —
// arrow aliases, whose hjkl twins already document them.
type binding struct {
	key  string
	desc string
	act  func(*editor)
}
