// view is an immutable snapshot of what to paint.
type view struct {
	cwd     string
	files   []entry
	idx     int
	sel     map[string]bool
	clip    int
	cutting bool
	alt     string
	msg     string
	w       int
	h       int
}
