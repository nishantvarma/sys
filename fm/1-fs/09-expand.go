func expand(s string) string {
	if s == "~" {
		return userHome
	}
	if strings.HasPrefix(s, "~/") {
		return filepath.Join(userHome, s[2:])
	}
	return s
}
