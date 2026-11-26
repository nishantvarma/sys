package outline

import (
	"fmt"

	"sys/tty"
)

// Help shows keys on a view that is no terminal, as a window is, in
// messages alone: a node a key, the key and what it does, which the view
// lays out as rows, and it waits for esc, enter, q or ? to close. A view
// keeps some keys, t among them, so not any key can close it
// (`todo/04-gio/02-term`); a move lands on a row and moves the cursor
// there. Each key is named as tty.Help names it on the terminal, and each
// row by its number under the title, so a view's head says title : row.
func Help(ui View, title string, keys [][2]string) {
	n := &Node{ID: title, Kind: "help", Text: title}
	w := 0
	for _, k := range keys {
		if k[1] != "" {
			w = max(w, len([]rune(tty.Label(k[0]))))
		}
	}
	for _, k := range keys {
		if k[1] == "" {
			continue
		}
		key := tty.Label(k[0])
		said := fmt.Sprintf("%-*s  %s", w, key, k[1])
		n.Kids = append(n.Kids, &Node{ID: At(title, len(n.Kids)+1),
			Kind: "key", Text: said})
	}
	if len(n.Kids) > 0 {
		n.Kids[0].Cur = true
	}
	for {
		ui.Draw(n, title+": esc closes", []string{"help"})
		k, ok := ui.Key()
		if !ok {
			return
		}
		switch k.Name {
		case tty.KEsc, tty.KEnter, "q", "?":
			return
		}
		for _, r := range n.Kids {
			if k.Node != "" {
				r.Cur = r.ID == k.Node
			}
		}
	}
}
