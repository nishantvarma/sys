func tilde(p string) string {
	if userHome != "" &&
		(p == userHome || strings.HasPrefix(p, userHome+"/")) {
		return "~" + p[len(userHome):]
	}
	return p
}
