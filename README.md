# sys

A collection of small, suckless-style tools for a programmer who wants a
system tailored to how they work, and the means to reshape it fast. Built
to be read and changed, not configured — there's no config format, so
making something behave differently is never more than a few lines and a
rebuild away.

## programs

    ed        modal editor, kakoune-flavoured: select first,
              then act on the selection; `ed -f` edits whatever cur
              names, and follows it while nothing is unsaved
    fm        single-pane file manager: vim keys, a remembered cursor per
              directory, and it shells out to your own tools instead of
              reimplementing them; `fm -f` goes where cur is said
    gio       sys's views in a window, drawn by gio: studio's frame, fm
              and ed in it as they are; its own module, built apart
              (`gio/README`)
    studio    the frame: fm, ed on the thing at fm's cursor, and a shell,
              as tmux panes or as a page; `studio/frame` is the layout, a
              note, and a program reaches each pane by its name, as the
              mouse's three buttons do (`design/mouse`); `studio show` is
              the lens, the thing at cur seen another way; run from a
              shell alone in its window it lays the frame out there, the
              shell its shell; q in fm ends the frame, asking first while
              ed holds work unsaved, and a shell that exits ends alone
    tangle    literate tangler: weaves a tree of code chunks into one main.go

and the packages they share:

    outline   the tree a view draws, and the messages between a program
              and its view; a note read into that tree by its indent
    tty       the shared terminal layer: raw mode, escape codes,
              key and click decoding, line input
    view      the outline drawn on a terminal: fm's layouts, folder and
              thumbs, on the whole screen or in a region a host gives,
              or as runs of text for a host that is no terminal, gio
    watch     says when files change in a folder: how a program follows
              cur
    state     a tmux window's own files, cur and the lens: where each is
              kept, read, and written whole
    figure    a note as its parts and the messages between them, laid
              out in cells: SVG on a page, runes on a terminal;
              `design/arch` is sys drawn that way, and `./build` makes
              it `arch.html`

ed, fm, studio and gio are what's built so far; the packages are what
they're built on, and tangle is how. Nothing here is a fixed scope — a window
manager, or whatever else earns its keep, can join the same way: small,
and hardcoded like the rest.

ed, fm and tmux each have a cheatsheet beside them — `ed/cheatsheet`,
`fm/cheatsheet`, `tmux/cheatsheet` — carrying the model, which is the half
no key table generates.

## the source is the interface

Keybindings, colours, and the programs each command spawns are ordinary Go
values, not a format to parse — see dependencies below for what's spawned,
and `sh/` for the scripts themselves. Want it different? Edit it and
rebuild — the same outcome as a config file, minus the file, the parser,
and the drift between them.

## the directory you are in

`o` opens it, `q` closes it. With no argument `o` takes cur, which is the
directory you are in until something is said, and a directory is what it
dispatches to fm. Nothing in `tmux/conf` names fm, so changing `sh/o`
changes what `o`, a right click and `` `p `` open. fm's `o` runs it too,
handing it the row under the cursor. `o` opens and never runs: a script
is text, and running is the shell's.

There is no key for this, because there is nothing a key would know that
the shell does not. fm splits in beside the shell and never moves it or
types at it, so there is nothing to keep in sync and nothing to keep
anywhere.

## the word you are on

`c` is the pronoun. Bare it prints cur — what the last verb, view or click
was about; with text it resolves that text to an address and keeps it. Every
verb takes its noun from `c`, so `o`, `e` and `p` with no argument mean this
file, `fm` with none opens on it, and everything that isn't ours takes
it as `` `{c} `` — ``git add `{c}``, ``wc -l `{c}``.

Everything that selects says it and nothing else does: fm's cursor as it
moves, and the lens's, a click moving either; a left click on a word in
any pane but theirs and ed's, which take the mouse for themselves; and ed
on save — with the line it saved at, so `e` reopens there. `ed -f` and
`fm -f` follow cur, and so a frame goes where anything says. One cur per
tmux window, one log for the machine: each line is a verb and the
address it was said about, so `m` walks that log newest first — back is
`` c `{m 1} ``, a line picked whole is that act again, and no view needs
a back key.

The mouse is acme's three buttons on these verbs, the same in every
pane: left is this, `c`; right is look, `p`, which opens what the word
names; middle is do, `p -a`, which hands it to the shell's prompt, where
Return runs it (`design/mouse`).

## the filesystem is the outline

`tangle` is literate programming with no literate-programming format. A program
is a tree of folders. Each chunk is a file `NN-name.go` holding one bare
declaration. `@main.go` is the root: `<<name>>` splices one chunk by name,
`<<others>>` sweeps the rest, in filesystem order. Folders nest like any
outline, and a chunk's name is scoped by its path — normalmode/key and
insertmode/key coexist, both under `ed/3-mode`; `<<name>>` takes any path
suffix that names one chunk. tangle walks the tree and stitches it into a
single `main.go`. Browse the outline in fm, edit the chunks in ed — no
markup to learn. `tangle`, `fm`, `ed` and `studio` are each written this
way; `outline`, `tty`, `view`, `watch`, `state` and `figure` stay plain
Go packages.

The outline does not stop at a file. A note is an outline too, by its
indent, so one tree runs from the root, the topmost thing, down through
folders and files to a single item in a note. Two panes work it: fm walks
the folders, a level at a time, and ed beside it edits the file at fm's
cursor, whose items are its indent. ed stays an editor: what more is
known of a file's structure is for something beside it to say.
outline.Text reads a note into that tree by its indent, as Python is
read, for what draws a note another way: studio's frame and its figure.

## everything talks in messages

The outline is what everything is; messages are how it talks, and the
one is as much a foundation as the other. A program and its view, and one
program and another, pass only messages: plain data, which crosses a
channel, a pipe or a socket alike. So any piece can be another process,
a view can be a terminal, a window or a page, and any program can drive
another by sending it an idea. Nothing is shared but the messages; what a
piece knows of another is what it was last told. `c` and the `m` log are
messages already, a verb and the address it was said about, passed
through files. The messages a program and its view pass are in
`design/view`.

## the studio, and where it starts

The studio is sys on one idea, in `design/studio`: a thing is a file or a
folder, a lens shows it, an id joins the lenses, the page in the browser
reads, and the terminal acts. `medium/` is the page side: `page.html`, the
one page every lens is drawn on; the renderers that make it from a note, a
folder, a python file or a diff, in `medium/page`; and the shelf that
reads them.

It starts at the todo, in `design/todo`: what you mean to do, as a tree of
files, in the order you do it, so the order is the priority and you are
the one who runs it. `todo` is its verb: bare it prints the walk of the
todo you are on, `todo -w` writes its page, and `todo -c` checks it. The
page and the walk are lenses on the folders, so neither is ever edited.

## rule

`rule/` is how the work is done — one rule per file, the name is the rule,
the body only if the name isn't enough. Same trick as tangle: the tree is
the outline, so read it in fm.

    general       simple, terse, symmetry, naming, root
    programming   decomposition, dry, stateless, module order
    learning      topdown, spacing, interleaving, connections
    debugging     reproduce, bisection
    ai            how to work with an assistant, in numbered order

None of it is compiled or read by any program, and little of it is about
sys — these are the rules first, and the programs are what came out of
applying them here.

## non-goals

Left out on purpose: marks, named registers, and an ex line in ed, whose
counts only replay a key; syntax highlighting; an fm preview pane or sort
modes; untangling main.go back into chunks. The tree is the source, and
the answers to the rest are your shell and your own tools.

## dependencies

Plan 9 is the foundation: `rc` is the shell fm shells out to and most of
`sh/` is written in, and `page` (Plan 9's image viewer) is one of `o`'s
openers. None of it is load-bearing on the Go side, though — rewrite `sh/`
in POSIX and fm won't notice; it only ever execs these by name.

    rc            shell fm shells out to, and most of sh/'s language
    st            terminal spawn opens when there is no tmux
    zoxide        directory-jump history behind fm's z
    xclip         X's selections: fm's and ed's copy, ed's paste, and
                  what tmux's copy mode selects
    fzy, fd, rg   fuzzy pick + search behind f, s and fm's picks
    ruff          linter/formatter behind lint, and fm's F
    node          prettier behind fmt, via npx; medium/shelf's and
                  medium/page's browser checks
    python3       medium/page/diff, a page from a diff.md and two trees
    tree, git     outline browsing, v's commit flow
    gitk          fm's git-gui binding, and one of v's own shortcuts
    tmux, wmctrl  panes and windows for spawn
    xdotool       the pointer warp moves
    zathura, mpv, mocp, foliate, xchm, microsoft-edge
                  o's openers, by extension, and edge p's for a url

`sh/` is the scripts these bindings actually run — c, m, p, e, o, v,
spawn, warp, todo, lint, check, the finders f and s, and v's small prompt
helpers. `./build` copies them into bin/ next to the compiled programs;
edit or replace them like anything else here.

`tmux/` is the same idea one layer out: the bindings that run those
scripts — a left click to say a word, a right click to open what it
names, a middle click to hand it to the shell — kept beside them rather
than in your dotfiles, so a trigger and its program can't drift apart.
The layout is here for the same reason: `spawn` asks for a pane and
never says where, so where one goes is a line of tmux and not a branch
in a script. Everything else about tmux stays yours; your own tmux.conf
reaches this with

    source-file ~/src/sys/tmux/conf

## build

    ./build          everything, into bin/
    ./build fm ed    named targets only
    gio/build        the window fm -w and ed -w open, apart, for it
                     needs cgo (`gio/README`)
    ./fmt            gofmt the tree, but tangle's sources, and prettier
                     every .html and .mjs

Each target is tangled, vetted and compiled, so `bin/tangle` must already
exist — it's tangled by itself. go's own tools read the tangled `main.go`
and never the chunks, which are no package: `go vet ./...` and gopls
choke on them, and `go vet main.go` in a target's folder is the way, as
`build` vets each. A `<<name>>` that names no chunk stops the tangle. `build` also copies `sh/` into `bin/`, so the
compiled programs and the scripts they call end up in one place — put
`bin/` on your `$PATH` and that's the whole environment. No install step,
no config — small enough to read in one sitting, and meant to stay that
way. `check` runs every check the tree ships, the packages' tests among
them.
