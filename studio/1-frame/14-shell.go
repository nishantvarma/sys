// shell is the frame's shell, the first slot that runs nothing; nil when
// every slot runs something.
func (f *frame) shell() *slot {
	for _, s := range leaves(f.slots) {
		if s.cmd == "" {
			return s
		}
	}
	return nil
}
