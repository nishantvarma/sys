package outline

import "testing"

// shown is a view that is no terminal, as a test holds one: what was
// drawn last, and the keys to hand back in turn.
type shown struct {
	last *Node
	keys []Key
}

func (s *shown) Draw(n *Node, status string, state []string) {
	s.last = n
}

func (s *shown) Ask(string, string) (string, bool) { return "", false }
func (s *shown) Size() (int, int)                  { return 80, 24 }
func (s *shown) Copy(string) error                 { return nil }

func (s *shown) Key() (Key, bool) {
	if len(s.keys) == 0 {
		return Key{}, false
	}
	k := s.keys[0]
	s.keys = s.keys[1:]
	return k, true
}

// Help draws a row a key, its key named as on the terminal and its words
// beside it, moves with a landing, and closes on esc and not on t.
func TestHelp(t *testing.T) {
	v := &shown{keys: []Key{{Node: At("ed — keys", 2)}, {Name: "t"},
		{Name: "ESC"}, {Name: "x"}}}
	Help(v, "ed — keys", [][2]string{{"h", "left"}, {"UP", ""},
		{" ", "collapse"}, {"INT", "copy"}})
	n := v.last
	if n == nil || len(n.Kids) != 3 ||
		n.Kids[1].Text != "space  collapse" ||
		n.Kids[2].Text != "C-c    copy" || !n.Kids[1].Cur {
		t.Fatalf("help drew %+v", n)
	}
	if len(v.keys) != 1 {
		t.Errorf("esc did not close the help: %d keys left",
			len(v.keys))
	}
}
