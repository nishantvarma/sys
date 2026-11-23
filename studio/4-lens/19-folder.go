// folder is where a change to the thing at path shows: the folder
// itself, or the folder a file is in.
func folder(path string, fi os.FileInfo) string {
	if fi.IsDir() {
		return path
	}
	return filepath.Dir(path)
}
