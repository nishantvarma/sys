func hl(n *outline.Node) string {
	switch {
	case n.Cur:
		return curBg
	case n.Sel:
		return selBg
	default:
		return ""
	}
}
