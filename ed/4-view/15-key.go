func (v *term) Key() (outline.Key, bool) {
	k, ok := v.t.ReadKey()
	return outline.Key{Name: k.Token(), Text: k.Text}, ok
}
