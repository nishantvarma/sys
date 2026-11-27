// read makes the frame from a note's items: the first is the title, the
// lines under it the lead (outline.Head), and every item after it a slot,
// in the order the slots take room (layout). A note with no slots is no
// frame, an error at the note.
func read(note string, items []*outline.Node) (*frame, error) {
	if len(items) < 2 {
		return nil, fmt.Errorf("%s: a frame is a title and slots",
			note)
	}
	title, lead := outline.Head(items)
	f := &frame{title: title, lead: lead}
	named := map[string]bool{}
	var slots []*slot
	for _, n := range items[1:] {
		s, err := slotOf(n, named)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	f.slots, f.panel = layout(slots)
	return f, nil
}
