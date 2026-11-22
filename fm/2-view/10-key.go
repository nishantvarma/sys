// Key reads the next key. What is the view's own it does itself: t turns
// to the next layout, the folder, the tree, tiles; a move goes down, up,
// to the last, a page, or to a row by its number, and as tiles right and
// left too; in the tree, a node opens and closes in place. fm hears only
// where the cursor landed, and what the view wants to see inside.
func (v *term) Key() (outline.Key, bool) {
	for {
		k, ok := v.t.ReadKey()
		if !ok {
			return outline.Key{}, false
		}
		name := k.Token()
		if name == "t" {
			next := map[string]string{"": "tree", "tree": "tiles"}
			v.lay = next[v.lay]
			v.show(v.status)
			continue
		}
		if to, moved := v.move(name); moved {
			return outline.Key{Node: to}, true
		}
		if ev, turned := v.turn(name); turned {
			return ev, true
		}
		if len(name) == 1 && name[0] >= '0' && name[0] <= '9' {
			return outline.Key{Node: v.jump(name)}, true
		}
		return outline.Key{Name: name, Text: k.Text}, true
	}
}
