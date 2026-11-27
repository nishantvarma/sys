// turn is what is asked of a slot's tabs. Bare, it says them, the one shown
// marked *; next and prev show the tab after or before the shown one,
// round, and a number the tab at that place, from 1; close ends the shown
// one, the next taking its place; a name shows the tab by that name,
// wherever it waits; new opens a terminal (newTab); and menu, given a
// client, opens tmux's own menu of them over the slot (menu). With no pane
// it runs in no tmux, and the frame's host is gio's window, which takes
// the ask from the window's state (ask).
func turn(slot, pane string, args []string) error {
	word := ""
	if len(args) > 0 {
		word = args[0]
	}
	if len(args) > 1 && word != "new" && word != "menu" {
		return errors.New("usage: studio tab [-t pane]" +
			" [next | prev | number | name | close |" +
			" menu [client] | new [command...]]")
	}
	if pane == "" {
		return ask(slot, word)
	}
	t := &tmux{}
	win, _ := t.run("display", "-p", "-t", pane, "#{@frame}")
	if win == "" {
		w, err := target()
		if err != nil {
			return err
		}
		win = w
	}
	if word == "new" {
		return t.newTab(win, strings.Join(args[1:], " "))
	}
	_, numbered := strconv.Atoi(word)
	if word != "" && numbered != nil && !slices.Contains([]string{"next",
		"prev", "close", "menu"}, word) {
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
	case word == "menu":
		client := ""
		if len(args) > 1 {
			client = args[1]
		}
		return t.menu(win, slot, client, tps, at)
	case numbered == nil:
		n, _ := strconv.Atoi(word)
		if n < 1 || n > len(tps) {
			return fmt.Errorf("no tab %d of %s: it has %d", n,
				slot, len(tps))
		}
		return t.show(win, slot, tps[n-1])
	case at < 0 && len(tps) > 0:
		// none shown: the panel, hidden, shows the tab it hid
		was, _ := t.run("show", "-wqv", "-t", win, "@studio-hidden")
		tp := tps[0]
		for _, o := range tps {
			if o.pane == was {
				tp = o
			}
		}
		return t.show(win, slot, tp)
	case at < 0:
		return nil // no tabs to show
	case word == "close":
		return t.closeTab(win, slot, tps, at)
	case len(tps) < 2:
		return nil // one tab: nothing to turn to
	}
	d := 1
	if word == "prev" {
		d = len(tps) - 1
	}
	return t.show(win, slot, tps[(at+d)%len(tps)])
}
