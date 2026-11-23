// age is how long ago, said in its largest whole unit: 3y, 2w, 5d, 7h,
// 12m, or now.
func age(d time.Duration) string {
	s := int64(d.Seconds())
	for _, u := range []struct {
		unit string
		n    int64
	}{{"y", 31536000}, {"w", 604800}, {"d", 86400}, {"h", 3600},
		{"m", 60}} {
		if s >= u.n {
			return fmt.Sprintf("%d%s", s/u.n, u.unit)
		}
	}
	return "now"
}
