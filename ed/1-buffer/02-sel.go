// sel is a range; head is the cursor. A collapsed sel (anchor == head) is a
// bare cursor — the one-rune degenerate case both models share.
type sel struct{ anchor, head pos }
