// words is what the view says of itself on the bottom line: its layout,
// and where the cursor is among what it shows, 12/34.
func (v *term) words() []string {
	lay := v.lay
	if lay == "" {
		lay = "folder"
	}
	if v.last == nil {
		return []string{lay}
	}
	_, rows := v.shown()
	if len(rows) == 0 {
		return []string{lay}
	}
	return []string{lay, fmt.Sprintf("%d/%d", curRow(rows)+1, len(rows))}
}
