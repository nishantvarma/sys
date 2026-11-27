// turn is what is asked of a frame's tabs, the words each host takes
// alike, l's and a press on a tab line's button (`design/lens`): bare, it
// says the slot's tabs, the one shown marked *; next and prev show the
// tab after or before the shown one, round; a number the tab at that
// place, from 1, as a tab line counts them; a name the tab by that name,
// in any slot of the frame; close ends the one shown, the next taking its
// place, but a slot's last tab stays, for it is the slot; and new opens a
// terminal in the panel, running a shell, or its words, one as a shell
// reads it, more as a command and its arguments, as tmux's new-window
// runs them (newTab). The eye goes to the tab shown, and a panel hidden
// shows again, with the tab it hid. With no pane it runs in no tmux, and
// the frame's host is gio's window, which the ask is put to (ask).
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
	if word == "new" {
		return t.newTab(win, pane, args[1:])
	}
	n, err := strconv.Atoi(word)
	place := err == nil
	if word != "" && !place && !slices.Contains([]string{"next", "prev",
		"close"}, word) {
		return t.showNamed(win, word)
	}
	if slot == "" {
		slot, _ = t.run("display", "-p", "-t", pane, "#{@slot}")
	}
	if slot == "" {
		return errors.New("not in a frame's slot: name the tab")
	}
	tps, err := t.tabsOf(win, slot)
	if err != nil {
		return err
	}
	at := slices.IndexFunc(tps, func(tp tabPane) bool {
		return tp.window == win
	})
	switch {
	case word == "":
		for i, tp := range tps {
			mark := " "
			if i == at {
				mark = "*"
			}
			fmt.Println(mark + " " + tp.name)
		}
		return nil
	case place && (n < 1 || n > len(tps)):
		return fmt.Errorf("no tab %d of %s: it has %d", n, slot,
			len(tps))
	case place:
		return t.show(win, slot, tps[n-1])
	case at < 0 && len(tps) > 0: // none shown: the panel, hidden
		return t.show(win, slot, t.hidden(win, tps))
	case at < 0:
		return nil // no tabs to show
	case len(tps) < 2 && word == "close":
		return fmt.Errorf("%s keeps its last tab: it is the slot",
			slot)
	case len(tps) < 2:
		return nil // one tab: nothing to turn to
	case word == "close":
		return t.closeTab(win, slot, tps, at)
	}
	d := 1
	if word == "prev" {
		d = len(tps) - 1
	}
	return t.show(win, slot, tps[(at+d)%len(tps)])
}
