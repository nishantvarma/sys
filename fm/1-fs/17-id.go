// id is a row's id: its folder's path and its name, kept as they are, so
// . and .. stay rows and are not taken for the folders they name.
func id(dir, name string) string {
	return strings.TrimSuffix(dir, "/") + "/" + name
}
