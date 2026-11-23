// size is n bytes said short: 640, 1.5K, 12K, 3.1M, a tenth kept only
// below ten.
func size(n int64) string {
	v := float64(n)
	for _, unit := range []string{"", "K", "M"} {
		if v < 1024 {
			if unit == "" || v >= 10 {
				return fmt.Sprintf("%d%s", int64(v), unit)
			}
			return fmt.Sprintf("%.1f%s", v, unit)
		}
		v /= 1024
	}
	return fmt.Sprintf("%.1fG", v)
}
