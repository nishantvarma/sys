// number is the number a name starts with, the order of a todo's nodes
// and of a tangle tree's chunks (`design/todo:/the NN- of a name/`): its
// digits and the dash after them, 02- of 02-taxes; empty when it has none.
func number(name string) string {
	i := 0
	for i < len(name) && name[i] >= '0' && name[i] <= '9' {
		i++
	}
	if i == 0 || i == len(name) || name[i] != '-' {
		return ""
	}
	return name[:i+1]
}
