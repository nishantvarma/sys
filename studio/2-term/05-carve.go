// carve splits pane among kids, down or across: the first kid keeps the
// pane, and each split hands the new pane the part of what is left that
// the kids after it take. Then each kid with slots of its own is carved
// the other way. -d leaves the first pane the one looked at, so fm,
// asking at its start whether it has the eye, hears yes.
func (t *tmux) carve(pane string, kids []*slot, across bool, dir string,
	panes map[*slot]string) error {
	w := weights(kids)
	way := "-v"
	if across {
		way = "-h"
	}
	at := pane
	for i, k := range kids {
		if i > 0 {
			pct := max(1, min(99, 100*sum(w[i:])/sum(w[i-1:])))
			p, err := t.run("split-window", "-d", way, "-t", at,
				"-l", strconv.Itoa(pct)+"%", "-c", dir,
				"-P", "-F", "#{pane_id}")
			if err != nil {
				return err
			}
			at = p
		}
		panes[k] = at
	}
	for _, k := range kids {
		if len(k.kids) == 0 {
			continue
		}
		if err := t.carve(panes[k], k.kids, !across, dir,
			panes); err != nil {
			return err
		}
	}
	return nil
}
