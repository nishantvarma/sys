// jump reads a row's number, the digits so far shown as the status, and
// commits once no longer number could stay in range: the id of the row it
// names, else empty.
func (v *term) jump(first string) string {
	buf := first
	_, rows := v.shown()
	n := len(rows)
	for {
		if x, _ := strconv.Atoi(buf); x*10 > n {
			break
		}
		v.show(buf)
		k, ok := v.t.ReadKey()
		name := k.Token()
		if !ok || name == tty.KEsc || name == tty.KInt {
			return ""
		}
		if name == tty.KEnter || len(name) != 1 || name[0] < '0' ||
			name[0] > '9' {
			break
		}
		buf += name
	}
	if x, err := strconv.Atoi(buf); err == nil && x >= 1 && x <= n {
		return rows[x-1].n.ID
	}
	return ""
}
