// mapLine replaces row l with f's result, handing f a clone so history's
// copy is untouched. This is the buffer's one rule — a line's runes are
// never written in place, only replaced — kept in the one place that knows
// it, instead of at each call site that has to remember. f may edit the
// clone it is given and return it; indent and dedent build a new line
// instead, and either is fine.
func (b *buffer) mapLine(l int, f func([]rune) []rune) {
	b.lines[l] = f(slices.Clone(b.lines[l]))
	b.dirty = true
}
