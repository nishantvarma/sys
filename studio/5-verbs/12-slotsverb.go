// slotsVerb prints a frame's slots as JSON, for a host that lays the frame
// out itself, as gio does: studio's own frame, or the note given. Each
// slot is its name, its weight among the slots beside it (weights), what
// runs in it, nothing for a shell, and the slots it is split into, the
// other way; the frame's own stack down.
func slotsVerb(args []string) error {
	if len(args) > 1 {
		return errors.New("usage: studio slots [frame]")
	}
	f, err := load(args)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(placedOf(f.slots))
}
