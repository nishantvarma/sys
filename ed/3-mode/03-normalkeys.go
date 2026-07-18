// normalKeys is the single source of truth: it drives dispatch AND the ?
// help screen, so the two can never drift. A vimNormal would carry its
// own table and reuse these same editor methods. It's a func, not a var,
// so help (which reads it) and the table (which names help) aren't a
// package init cycle.
//
// Repeats re-aim: w w w is three fresh hops, not one growing selection.
// Shape verbs absorb repeats as growth — x takes the next line, %
// already has it all — and v turns re-aiming into growing for every
// motion at once: kakoune's shifted row, one key. A numeric prefix is
// only that repeat spelled out: 3w is w w w, dispatch replaying the
// key, so nothing here needs to know a count exists.
func normalKeys() []binding {
	return []binding{
		{"h", "left", (*editor).left},
		{"j", "down", (*editor).down},
		{"k", "up", (*editor).up},
		{"l", "right", (*editor).right},
		{tty.KLeft, "", (*editor).left},
		{tty.KDown, "", (*editor).down},
		{tty.KUp, "", (*editor).up},
		{tty.KRight, "", (*editor).right},
		{"0", "line start", (*editor).toStart},
		{"$", "line end", (*editor).toEnd},
		{"g", "first line", (*editor).toTop},
		{"G", "last line", (*editor).toBottom},
		{tty.KPgDn, "page down", (*editor).pageDown},
		{tty.KPgUp, "page up", (*editor).pageUp},
		{":", "goto line", (*editor).goToLine},
		{"w", "word +space", (*editor).selWord},
		{"b", "word back", (*editor).selWordBack},
		{"e", "word end", (*editor).selWordEnd},
		{"m", "match bracket", (*editor).selMatch},
		{"{", "prev blank", (*editor).selParaBack},
		{"}", "next blank", (*editor).selParaFwd},
		{"f", "find char", (*editor).findFwd},
		{"t", "till char", (*editor).tillFwd},
		{"F", "find back", (*editor).findBack},
		{"T", "till back", (*editor).tillBack},
		{"/", "search", (*editor).search},
		{"n", "next match", (*editor).searchNext},
		{"N", "prev match", (*editor).searchPrev},
		{"*", "search sel", (*editor).searchSel},
		{"x", "line (xx grows)", (*editor).selectLine},
		{"%", "all", (*editor).selectAll},
		{"v", "extend", (*editor).extend},
		{" ", "collapse", (*editor).collapse},
		{"i", "insert", (*editor).insertBefore},
		{"a", "append", (*editor).insertAfter},
		{"A", "append eol", (*editor).insertEol},
		{"o", "open below", (*editor).openBelow},
		{"O", "open above", (*editor).openAbove},
		{"d", "delete", (*editor).del},
		{"c", "change", (*editor).change},
		{"r", "replace", (*editor).replaceChar},
		{"J", "join", (*editor).joinLines},
		{">", "indent", (*editor).indent},
		{"<", "dedent", (*editor).dedent},
		{"y", "yank +clipboard", (*editor).yank},
		{"p", "paste, replace sel", (*editor).paste},
		{"P", "paste before", (*editor).pasteBefore},
		{"u", "undo", (*editor).undo},
		{"U", "redo", (*editor).redo},
		{"s", "save", (*editor).doSave},
		{"R", "reload", (*editor).reload},
		{"q", "quit (2× discard)", (*editor).quit},
		{"?", "help", (*editor).help},
	}
}
