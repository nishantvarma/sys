// asked is an ask as the window's ask file holds it, a line of JSON,
// which gio reads back (`gio/host.go:/type asked struct/`): the verb, tab
// or panel, the slot asked of, "." for the one with the eye, and the
// verb's words.
type asked struct {
	Verb  string   `json:"verb"`
	Slot  string   `json:"slot,omitempty"`
	Words []string `json:"words,omitempty"`
}
