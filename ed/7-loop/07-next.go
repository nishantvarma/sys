// next is the next key: one that . put back, ahead of anything typed, else
// the view's. A put-back key comes the way a typed one does, by its token,
// so the two reach the mode alike.
func (e *editor) next() (outline.Key, bool) {
	if len(e.fed) > 0 {
		k := e.fed[0]
		e.fed = e.fed[1:]
		return outline.Key{Name: k.Token(), Text: k.Text}, true
	}
	return e.ui.Key()
}
