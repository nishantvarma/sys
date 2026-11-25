// follow goes where cur is said, with -f, as ed -f does in a frame's slot:
// cur is a row, so fm stands in the folder that holds it and lands on it,
// a folder too, as fm with no argument opens on cur (here), and a file's
// line rides along for e (hitLine). cur naming the row already, by any
// path, moves nothing, and neither does cur that names nothing there is.
// fm's own word comes back to it as cur too, and names its row.
func (m *fm) follow() {
	at := state.Read(m.this)
	if !m.follows || at == "" || at == m.heard {
		return
	}
	m.heard = at
	p, line := outline.Address(at)
	if same(p, m.cur()) || !exists(p) {
		return
	}
	if d := filepath.Dir(p); d != m.cwd {
		if m.cd(d); m.cwd != d {
			return
		}
	}
	m.land(filepath.Base(p))
	if m.cur() == p {
		m.hit, m.line = p, line
	}
}
