// shell is the frame's shell, its panel, the terminals: the first slot,
// as the note says them, that runs nothing (layout); nil when every slot
// runs something.
func (f *frame) shell() *slot {
	return f.panel
}
