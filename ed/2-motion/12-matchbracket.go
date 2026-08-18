// matchBracket lands on the bracket that pairs with the one under p (by
// nesting depth), or p itself if p isn't on a bracket or no match balances.
func matchBracket(b *buffer, p pos) pos {
	r, ok := runeAt(b, p)
	if !ok {
		return p
	}
	var other rune
	var step func(*buffer, pos) pos
	switch r {
	case '(':
		other, step = ')', fwd
	case ')':
		other, step = '(', back
	case '[':
		other, step = ']', fwd
	case ']':
		other, step = '[', back
	case '{':
		other, step = '}', fwd
	case '}':
		other, step = '{', back
	default:
		return p
	}
	depth := 1
	for {
		n := step(b, p)
		if n == p {
			return p // unbalanced; give up in place
		}
		p = n
		switch c, _ := runeAt(b, p); c {
		case r:
			depth++
		case other:
			depth--
			if depth == 0 {
				return p
			}
		}
	}
}
