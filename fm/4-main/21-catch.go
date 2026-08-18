// catch flashes an error and says whether there was none, so a verb that
// made something can land on it.
func (m *fm) catch(err error) bool {
	if err != nil {
		m.flash("error: " + err.Error())
		return false
	}
	return true
}
