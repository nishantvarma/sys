// placed is a slot as a host reads it (slotsVerb).
type placed struct {
	Name   string   `json:"name"`
	Weight int      `json:"weight"`
	Run    string   `json:"run,omitempty"`
	Kids   []placed `json:"kids,omitempty"`
}
