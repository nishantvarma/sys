# what changed, in a line
what the change is and why, a paragraph or more. `code` is code. the map
below is optional: a tree of the change, indented four spaces and drawn
with ├ and └, a note after two spaces. a file the change adds shows blue.

    src/
    ├── a.py      what changed in it
    └── b.py      new: a file the change adds

old: path/to/old
new: path/to/new

## a node: one part of the change, named by what it does
why, as prose. a node holds leaves, then nodes one # deeper.
- a.py f → a.py f — a function, old against new; the note says why
- a.py f → b.py g — moved, and renamed
- - → b.py h — new: nothing on the old side
- a.py k → - — removed: nothing on the new side
- a.py → a.py — a whole file

### a node inside it
a side names a file, or a file and a symbol in it: a function, a class, an
assignment, `Class.member`, or `<head>`, what comes before the first def.
- a.py <head> → a.py <head> — the imports
