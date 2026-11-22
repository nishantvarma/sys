// move is the id of the row a key moves the cursor to, among those shown,
// empty when there are none; moved is false when the key is no move. As
// tiles, down and up go a line of tiles, and right and left one tile.
func (v *term) move(name string) (to string, moved bool) {
	if v.last == nil {
		return "", false
	}
	_, rows := v.shown()
	w, h := v.t.Size()
	step := 1
	if v.lay == "tiles" {
		step, _ = grid(rows, w)
	}
	d := 0
	switch name {
	case "j", tty.KDown:
		d = step
	case "k", tty.KUp:
		d = -step
	case tty.KPgDn:
		d = (h - 3) * step
	case tty.KPgUp:
		d = -(h - 3) * step
	case "G":
		d = len(rows)
	case tty.KRight, tty.KLeft:
		if v.lay != "tiles" {
			return "", false // in and out, fm's or the tree's
		}
		d = 1
		if name == tty.KLeft {
			d = -1
		}
	default:
		return "", false
	}
	if len(rows) == 0 {
		return "", true
	}
	return rows[max(0, min(curRow(rows)+d, len(rows)-1))].n.ID, true
}
