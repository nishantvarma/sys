// ctx is the working directory for spawned commands: a tagged symlink's
// target when browsing the tag dir, else the cwd.
func (m *fm) ctx() string {
	if m.cwd == m.tags {
		if c := m.cur(); c != "" && isDir(c) {
			return c
		}
	}
	return m.cwd
}
