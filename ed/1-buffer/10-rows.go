// rows, last, line and lineLen are the whole read side of the buffer.
// Everything above this folder asks in these terms — how many lines, and
// what is on one — and never names lines itself, so what a line is stored
// in stays a question only this folder answers.
func (b *buffer) rows() int { return len(b.lines) }
