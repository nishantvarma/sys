// placed is a slot as a host reads it (slotsVerb): its name, its weight
// among the slots beside it, what runs in it, the slots it is split into,
// when it has more than one, its tabs, each a name and what runs in it,
// the first the one shown, and whether it is the frame's panel, its
// terminals across the bottom, which a host hides and shows (layout).
type placed struct {
	Name   string   `json:"name"`
	Weight int      `json:"weight"`
	Run    string   `json:"run,omitempty"`
	Kids   []placed `json:"kids,omitempty"`
	Tabs   []placed `json:"tabs,omitempty"`
	Panel  bool     `json:"panel,omitempty"`
}
