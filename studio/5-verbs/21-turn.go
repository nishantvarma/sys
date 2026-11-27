// turn is what is asked of a frame's tabs, the words each host takes
// alike, l's and a press on a tab line's button: what each means is
// tabs.Turn's (`tabs/tabs.go:/func Turn/`), carried out here in tmux, the eye
// going to the tab shown (show), a terminal the panel's, opened in the
// folder of the pane asked from (newTab). Bare, it says the slot's tabs,
// the one shown marked *. With no pane it runs in no tmux, and the
// frame's host is gio's window, which the ask is put to (ask).
func turn(slot, pane string, args []string) error {
	word := ""
	if len(args) > 0 {
		word = args[0]
	}
	if len(args) > 1 && word != "new" {
		return errors.New("usage: studio tab [-t pane]" +
			" [next | prev | number | name | close |" +
			" new [command...]]")
	}
	if pane == "" {
		if word == "" {
			return errors.New("the window says its tabs on its" +
				" tab line")
		}
		return ask("tab", slot, args)
	}
	t := &tmux{}
	win, _ := t.run("display", "-p", "-t", pane, "#{@frame}")
	if win == "" {
		return errors.New("not in a frame")
	}
	if slot == "" {
		slot, _ = t.run("display", "-p", "-t", pane, "#{@slot}")
	}
	slots, panes, err := t.slotsOf(win)
	if err != nil {
		return err
	}
	in := slices.IndexFunc(slots, func(s tabs.Slot) bool {
		return s.Name == slot
	})
	if word == "" {
		if in < 0 {
			return errors.New("not in a frame's slot")
		}
		s := slots[in]
		for i, name := range s.Tabs {
			mark := " "
			if i == s.Shown && !s.Hidden {
				mark = "*"
			}
			fmt.Println(mark + " " + name)
		}
		return nil
	}
	act, err := tabs.Turn(slots, slot, args)
	if err != nil {
		return err
	}
	switch act.Do {
	case tabs.New:
		return t.newTab(win, pane, act.Run)
	case tabs.Show:
		return t.show(win, act.Slot, panes[act.Slot][act.Tab])
	case tabs.Close:
		return t.closeTab(win, act.Slot, panes[act.Slot],
			slots[in].Shown)
	}
	return nil
}
