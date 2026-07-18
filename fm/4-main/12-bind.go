func (m *fm) bind() {
	m.keys = map[string]*binding{}
	b := func(key, desc string, act func()) {
		bd := &binding{key, desc, act}
		m.keys[key] = bd
		m.order = append(m.order, bd)
	}
	b("a", "sessions", func() { m.detach(cmdAI) })
	b("b", "duplicate", func() { m.spawn(false, false, cmdOpen, m.cwd) })
	b("c", "copy", m.copy)
	b("C", "cut", m.cut)
	b("d", "delete", m.rm)
	b("e", "edit", m.edit)
	b("f", "check", func() { m.spawn(true, true, "lint") })
	b("F", "format", func() { m.spawn(true, true, "lint", "-f") })
	b("g", "fuzzy edit", func() { m.spawn(false, false, cmdFzE) })
	b("G", "last file", func() { m.mv(len(m.files)) })
	b("i", "add file", func() { m.create("touch", touch) })
	b(
		"I",
		"add dir",
		func() {
			m.create("mkdir", func(p string) error {
				return os.Mkdir(p, 0o755)
			})
		},
	)
	b("h", "", func() { m.cd(filepath.Dir(m.cwd)) })
	b("j", "", func() { m.mv(1) })
	b("k", "", func() { m.mv(-1) })
	b("l", "link", m.link)
	b("m", "clone", func() { m.clone(false) })
	b("M", "clone dir", func() { m.clone(true) })
	b("n", "next", func() { m.find(1) })
	b("N", "prev", func() { m.find(-1) })
	b("o", "outline", func() { m.spawn(true, false, "outline") })
	b("p", "paste", m.paste)
	b("q", "quit", func() { m.done = true })
	b("r", "rename", m.rename)
	b("s", "shell", m.sh)
	b("t", "tag", m.tag)
	b("u", "unclip", func() { m.clip = nil })
	b("v", "vc", func() { m.spawn(false, false, cmdVC) })
	b("V", "git gui", func() { m.spawn(false, false, "gitk") })
	b("w", "workflow", m.workflow)
	b("x", "fuzzy open", func() { m.spawn(false, false, cmdFzO) })
	b("y", "yank path", func() { m.yank(false) })
	b("Y", "yank content", func() { m.yank(true) })
	b("z", "zoxide", m.zoxide)
	b("Z", "goto", m.goto_)
	b(" ", "select", m.toggleSel)
	b("?", "help", m.help)
	b("*", "chmod +x", m.chmod)
	b(".", "hidden", func() { m.hidden = !m.hidden })
	b("/", "search", m.search)
	b("'", "alt", func() {
		if m.alt != "" {
			m.alt = ""
		} else {
			m.alt = m.cwd
		}
	})
	b(";", "grep", m.fzsearch)
	b("`", "tagged", func() { m.cd(m.tags) })
	b("~", "home", func() { m.cd(userHome) })
	b(tty.KEnter, "", m.enter)
	b(tty.KRight, "", m.enter)
	b(tty.KLeft, "", func() { m.cd(filepath.Dir(m.cwd)) })
	b(tty.KUp, "", func() { m.mv(-1) })
	b(tty.KDown, "", func() { m.mv(1) })
	b(tty.KTab, "", m.tab)
	b(tty.KPgDn, "page down", func() { _, h := m.t.Size(); m.mv(h - 3) })
	b(tty.KPgUp, "page up", func() { _, h := m.t.Size(); m.mv(-(h - 3)) })
}
