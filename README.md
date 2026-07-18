# sys

A collection of small, suckless-style tools for a programmer who wants a
system tailored to how they work, and the means to reshape it fast. Built
to be read and changed, not configured — there's no config format, so
making something behave differently is never more than a few lines and a
rebuild away.

## programs

    ed        modal editor, kakoune-flavoured: select first,
              then act on the selection
    fm        single-pane file manager: vim keys, a remembered cursor per
              directory, and it shells out to your own tools instead of
              reimplementing them
    tty       the shared terminal layer: raw mode, escape codes,
              key decoding, line input
    tangle    literate tangler: weaves a tree of code chunks into one main.go

ed and fm are what's built so far; tty and tangle are what they're built
on. Nothing here is a fixed scope — a window manager, or whatever else
earns its keep, can join the same way: small, and hardcoded like the rest.

## the source is the interface

Keybindings, colours, and the programs each command spawns are ordinary Go
values, not a format to parse — see dependencies below for what's spawned,
and `sh/` for the scripts themselves. Want it different? Edit it and
rebuild — the same outcome as a config file, minus the file, the parser,
and the drift between them.

## the filesystem is the outline

`tangle` is literate programming with no literate-programming format. A program
is a tree of folders. Each chunk is a file `NN-name.go` holding one bare
declaration. `@main.go` is the root: `<<name>>` splices one chunk by name,
`<<others>>` sweeps the rest, in filesystem order. Folders nest like any
outline, and a chunk's name is scoped by its path — `normalmode/key` and
`insertmode/key` coexist; `<<name>>` takes any path suffix that names one
chunk. tangle walks the tree and stitches it into a single `main.go`. Browse
the outline in fm, edit the chunks in ed — no markup to learn. `tangle`,
`fm`, and `ed` are each written this way; `tty` stays a plain Go package.

## rule

`rule/` is how the work is done — one rule per file, the name is the rule,
the body only if the name isn't enough. Same trick as tangle: the tree is
the outline, so read it in fm.

    general       simplicity, symmetry, naming, reduce, root
    programming   decomposition, dry, terse, stateless, module order
    learning      topdown, spacing, interleaving, connections
    debugging     reproduce, bisection
    ai            how to work with an assistant, in numbered order

None of it is compiled or read by any program, and little of it is about
sys — these are the rules first, and the programs are what came out of
applying them here.

## non-goals

Left out on purpose: motion counts, marks, named registers, and an ex line in
ed; syntax highlighting; an fm preview pane, sort modes, or bookmarks beyond
tags; untangling main.go back into chunks. The tree is the source, and the
answers to the rest are your shell and your own tools.

## dependencies

Plan 9 is the foundation: `rc` is the shell fm shells out to and most of
`sh/` is written in, and `page` (Plan 9's image viewer) is one of `o`'s
openers. None of it is load-bearing on the Go side, though — point
`shellCmd` at `bash`, rewrite `sh/` in POSIX, and fm won't notice; it only
ever execs these by name.

    rc            shell fm shells out to, and most of sh/'s language
    st            terminal fm spawns detached commands into
    zoxide        directory-jump history behind fm's z
    xclip         clipboard behind fm's yank
    fzy, fd, rg   fuzzy pick + search behind fze/fzo/fzs
    ruff          linter/formatter behind fm's f/F
    tree, git     outline browsing, v's commit flow
    gitk          fm's git-gui binding, and one of v's own shortcuts
    tmux, wmctrl  window and session handling for spawn

`sh/` is the scripts these bindings actually run — e, o, v, spawn, the
fuzzy finders, and v's small prompt helpers. `./build` copies them into
bin/ next to the compiled programs; edit or replace them like anything
else here.

## build

    ./build          everything, into bin/
    ./build fm ed    named targets only
    ./fmt            gofmt the tree (skips tangle sources)

Each target is tangled then compiled, so `bin/tangle` must already exist —
it's tangled by itself. `build` also copies `sh/` into `bin/`, so the
compiled programs and the scripts they call end up in one place — put
`bin/` on your `$PATH` and that's the whole environment. No install step,
no tests, no config — small enough to read in one sitting, and meant to
stay that way.
