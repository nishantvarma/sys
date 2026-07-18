func (m *fm) catch(err error) {
	if err != nil {
		m.flash("error: " + err.Error())
	}
}
