// tab is one program that takes its turn in a slot: its name, which a
// program reaches it by and the slot's bar shows; what runs in it, empty
// for a shell; and where the frame says it.
type tab struct {
	name, cmd, at string
}
