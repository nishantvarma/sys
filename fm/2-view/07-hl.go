func hl(i int, f entry, v view) string {
	switch {
	case i == v.idx:
		return curBg
	case v.sel[f.path]:
		return selBg
	default:
		return ""
	}
}
