// share reads a leading percent off the rest of a slot's line, such as
// 40%; what follows it is what runs in the slot. false when the first
// word ends in % and is no whole percent from 1 to 100.
func share(rest string) (int, string, bool) {
	lead, more, _ := strings.Cut(rest, " ")
	p, ok := strings.CutSuffix(lead, "%")
	if !ok {
		return 0, rest, true
	}
	n, err := strconv.Atoi(p)
	if err != nil || n < 1 || n > 100 {
		return 0, "", false
	}
	return n, strings.TrimSpace(more), true
}
