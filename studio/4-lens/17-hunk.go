// hunk is the first line of the file as it is now that a hunk covers:
// the c of @@ -a,b +c,d @@, 0 when the header is not one.
func hunk(l string) int {
	_, after, ok := strings.Cut(l, " +")
	if !ok {
		return 0
	}
	num, _, _ := strings.Cut(after, ",")
	num, _, _ = strings.Cut(num, " ")
	n, _ := strconv.Atoi(num)
	return n
}
