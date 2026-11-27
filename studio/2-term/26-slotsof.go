// slotsOf is a frame's slots as the tab words see them (tabs.Turn): each
// slot's tabs, wherever they wait (tabsOf), and the one in its place; and
// the panes they are in, by slot. A slot none of whose tabs is in the
// window is the panel, hidden, its place the tab it hid (hidden).
func (t *tmux) slotsOf(win string) ([]tabs.Slot, map[string][]tabPane,
	error) {
	out, err := t.run("list-panes", "-a", "-f",
		"#{==:#{@frame},"+win+"}", "-F", "#{@slot}")
	if err != nil {
		return nil, nil, err
	}
	var slots []tabs.Slot
	panes := map[string][]tabPane{}
	for _, name := range strings.Fields(out) {
		if _, seen := panes[name]; seen {
			continue
		}
		tps, err := t.tabsOf(win, name)
		if err != nil {
			return nil, nil, err
		}
		panes[name] = tps
		s := tabs.Slot{Name: name}
		for _, tp := range tps {
			s.Tabs = append(s.Tabs, tp.name)
		}
		s.Shown = slices.IndexFunc(tps, func(tp tabPane) bool {
			return tp.window == win
		})
		if s.Shown < 0 && len(tps) > 0 {
			s.Hidden, s.Shown = true, slices.Index(tps,
				t.hidden(win, tps))
		}
		slots = append(slots, s)
	}
	return slots, panes, nil
}
