func (p pos) less(q pos) bool {
	return p.line < q.line || (p.line == q.line && p.col < q.col)
}
