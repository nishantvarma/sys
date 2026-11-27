// thing is what the lens is on: the address cur names, the path in it,
// and the note it names (outline.Note), a file or a folder's README, with
// its text, ed's while ed holds it unsaved (noted) and the file's else.
// With no note, why says why: a folder with no README, a file gone, or
// one that is no note a person types.
type thing struct {
	at, path string
	note     string
	text     []byte
	held     bool
	why      error
}
