// Package tabs is what the words that turn a frame's tabs mean, l's and a
// press's on a tab line (`design/lens`): one meaning, which each host of
// the frame, studio in tmux and gio's window, asks of Turn and carries out
// its own way. A host says what its frame holds; Turn says what to do.
package tabs

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// Slot is a frame's slot as the words see it: its name; its tabs, by name,
// in their order; the one in its place, by its place from 0; and whether
// the slot is out of sight, as the panel hidden is, its place the tab it
// hid.
type Slot struct {
	Name   string
	Tabs   []string
	Shown  int
	Hidden bool
}

// Do is what a word asks a host to do.
type Do int

const (
	None  Do = iota
	Show     // the tab at Act.Tab of Act.Slot, its slot shown with it
	New      // a terminal in the frame's panel, running Act.Run
	Close    // the tab Act.Slot shows, the next taking its place
)

// Act is a word's ask of a host: what to do, to which slot and tab, and
// what a new terminal runs, a shell for none.
type Act struct {
	Do   Do
	Slot string
	Tab  int
	Run  []string
}

// Turn is what words ask of a frame's slots, said in the slot named at:
// next and prev, the tab after or before the one shown, round; a number,
// the tab at that place, from 1, as a tab line counts them; close, the
// one shown, but a slot's last tab stays, for it is the slot; new, a
// terminal running the words after it; and any other word the tab by that
// name, in any slot. A slot hidden shows again the tab it hid, whatever
// turns it, but a number, which names a place.
func Turn(slots []Slot, at string, words []string) (Act, error) {
	if len(words) == 0 {
		return Act{}, errors.New("no word")
	}
	word := words[0]
	if word == "new" {
		return Act{Do: New, Run: words[1:]}, nil
	}
	n, err := strconv.Atoi(word)
	place := err == nil
	if !place && !slices.Contains([]string{"next", "prev", "close"},
		word) {
		for _, s := range slots {
			if i := slices.Index(s.Tabs, word); i >= 0 {
				return Act{Do: Show, Slot: s.Name, Tab: i}, nil
			}
		}
		return Act{}, fmt.Errorf("no tab %s", word)
	}
	i := slices.IndexFunc(slots, func(s Slot) bool { return s.Name == at })
	if i < 0 {
		return Act{}, errors.New("not in a frame's slot: name the tab")
	}
	s := slots[i]
	switch {
	case place && (n < 1 || n > len(s.Tabs)):
		return Act{}, fmt.Errorf("no tab %d of %s: it has %d", n,
			s.Name, len(s.Tabs))
	case place:
		return Act{Do: Show, Slot: s.Name, Tab: n - 1}, nil
	case len(s.Tabs) == 0:
		return Act{}, nil
	case s.Hidden:
		return Act{Do: Show, Slot: s.Name, Tab: s.Shown}, nil
	case word == "close" && len(s.Tabs) < 2:
		return Act{}, fmt.Errorf("%s keeps its last tab: it is"+
			" the slot", s.Name)
	case word == "close":
		return Act{Do: Close, Slot: s.Name}, nil
	case len(s.Tabs) < 2:
		return Act{}, nil // one tab: nothing to turn to
	}
	d := 1
	if word == "prev" {
		d = len(s.Tabs) - 1
	}
	next := (s.Shown + d) % len(s.Tabs)
	return Act{Do: Show, Slot: s.Name, Tab: next}, nil
}
