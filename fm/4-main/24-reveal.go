// reveal puts the cursor on p, going to its directory first. A content search
// brings a line with it, which waits on hit for the next edit there.
func (m *fm) reveal(p string, line int) {
	if !filepath.IsAbs(p) {
		p = filepath.Join(m.ctx(), p)
	}
	if isDir(p) {
		m.cd(p)
		return
	}
	d := filepath.Dir(p)
	m.cd(d)
	if m.cwd == d {
		m.land(filepath.Base(p))
		m.hit, m.line = p, line
	}
}
