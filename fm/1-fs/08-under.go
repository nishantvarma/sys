// under reports whether p lies within root (root=="/" always matches).
func under(p, root string) bool {
	return root == "/" || p == root || strings.HasPrefix(p, root+"/")
}
