package outline

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"strings"
)

// Go reads Go source as its declarations, as Parts reads a note, a part
// each, of a kind a view draws it by:
//
//	title  the package clause, its doc under it
//	type   a type, its fields and its methods under it
//	func   a func, a method, or an interface's method
//	value  a const, a var, or a struct's field
//
// A part's text is its own line as written, a func's to its body, the
// name it declares a run of kind name, and its doc comment is a note,
// read as one (Parts), its first part what is said of it. A group,
// const ( … ), is a part of its own, its word its text and its specs
// under it. A tangle chunk, a declaration with no package clause
// (`README.md:/the filesystem is the outline/`), reads as one. Each part's
// id is id:n, n the line it starts on, its doc's first when it has one.
// Source that does not parse still reads as far as it does, and the error
// says where it stops, as the file numbers its lines.
func Go(r io.Reader, id string) ([]*Node, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	fs, chunk := token.NewFileSet(), 0
	if _, err := parser.ParseFile(token.NewFileSet(), id, src,
		parser.PackageClauseOnly); err != nil {
		chunk, src = 1, append([]byte("package chunk\n"), src...)
	}
	f, err := parser.ParseFile(fs, id, src, parser.ParseComments)
	if f == nil {
		return nil, err
	}
	if list, ok := err.(scanner.ErrorList); ok && len(list) > 0 {
		e := list[0]
		err = fmt.Errorf("%s:%d: %s", id, e.Pos.Line-chunk, e.Msg)
	}
	g := &golang{fs: fs, src: src, id: id, off: chunk}
	var out []*Node
	if chunk == 0 {
		title := "package " + f.Name.Name
		out = append(out, g.part("title", title, f.Package, f.Doc))
	}
	types := map[string]*Node{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}
			to := &out // where its specs go: under it, in a group
			if d.Lparen.IsValid() {
				grp := g.group(d)
				out, to = append(out, grp), &grp.Kids
			}
			for _, s := range d.Specs {
				n := g.spec(d, s)
				*to = append(*to, n)
				if t, ok := s.(*ast.TypeSpec); ok {
					types[t.Name.Name] = n
				}
			}
		case *ast.FuncDecl:
			sig := g.text(d.Pos(), body(d))
			n := g.part("func", sig, d.Pos(), d.Doc)
			g.named(n, d.Pos(), d.Name)
			if t := types[recv(d)]; t != nil {
				t.Kids = append(t.Kids, n)
			} else {
				out = append(out, n)
			}
		}
	}
	return out, err
}

// golang is a Go file as Go reads it: its positions and its source, the
// id its parts are named by, and how many lines were put before it, one
// for a chunk's package clause.
type golang struct {
	fs  *token.FileSet
	src []byte
	id  string
	off int
}

// group is a declaration that groups its specs, const ( … ), as its part:
// its word its text and its doc under it, its specs to come.
func (g *golang) group(d *ast.GenDecl) *Node {
	kind := "value"
	if d.Tok == token.TYPE {
		kind = "type"
	}
	return g.part(kind, d.Tok.String(), d.Pos(), d.Doc)
}

// spec is a type or a value a declaration holds, as its part: standing
// alone, from its word, type or const, else from its name, under its
// group; its doc its own, else, alone, its declaration's, else the
// comment after it.
func (g *golang) spec(d *ast.GenDecl, s ast.Spec) *Node {
	from, doc := d.Pos(), d.Doc
	if d.Lparen.IsValid() {
		from, doc = s.Pos(), nil
	}
	switch s := s.(type) {
	case *ast.TypeSpec:
		n := g.part("type", g.text(from, opening(s)), from,
			firstOf(s.Doc, doc, s.Comment))
		g.named(n, from, s.Name)
		switch t := s.Type.(type) {
		case *ast.StructType:
			for _, f := range t.Fields.List {
				n.Kids = append(n.Kids, g.field("value", f))
			}
		case *ast.InterfaceType:
			for _, f := range t.Methods.List {
				n.Kids = append(n.Kids, g.field("func", f))
			}
		}
		return n
	case *ast.ValueSpec:
		end := s.End()
		if g.fs.Position(from).Line != g.fs.Position(end).Line {
			end = g.lineEnd(from)
		}
		n := g.part("value", g.text(from, end), from,
			firstOf(s.Doc, doc, s.Comment))
		g.named(n, from, s.Names[0])
		return n
	}
	return nil
}

// firstOf is the first of the comments given that there is.
func firstOf(cs ...*ast.CommentGroup) *ast.CommentGroup {
	for _, c := range cs {
		if c != nil {
			return c
		}
	}
	return nil
}

// field is a struct's field or an interface's method as its part: its
// names and its type, its tag left out, and its doc, or the comment after
// it.
func (g *golang) field(kind string, f *ast.Field) *Node {
	n := g.part(kind, g.text(f.Pos(), f.Type.End()), f.Pos(),
		firstOf(f.Doc, f.Comment))
	if len(f.Names) > 0 {
		g.named(n, f.Pos(), f.Names[0])
	}
	return n
}

// named puts the runs of a part's text first under it: the name it
// declares, of kind name, and the text around it; none when its text,
// as written, does not hold the name where its source does.
func (g *golang) named(n *Node, from token.Pos, name *ast.Ident) {
	before := g.text(from, name.Pos())
	if before != "" {
		before += " "
	}
	rest, ok := strings.CutPrefix(n.Text, before+name.Name)
	if !ok {
		return
	}
	var runs []*Node
	for _, r := range []*Node{{Text: before},
		{Kind: "name", Text: name.Name}, {Text: rest}} {
		if r.Text != "" {
			runs = append(runs, r)
		}
	}
	n.Kids = append(runs, n.Kids...)
}

// part is a declaration's part, of kind, saying text, from pos: its doc
// read as a note, first under it.
func (g *golang) part(kind, text string, pos token.Pos,
	doc *ast.CommentGroup) *Node {
	n := &Node{ID: g.at(pos), Kind: kind, Text: text}
	if doc == nil {
		return n
	}
	n.ID = g.at(doc.Pos())
	var ls []line
	for _, c := range doc.List {
		at := g.fs.Position(c.Pos()).Line - g.off
		for i, t := range strings.Split(comment(c.Text), "\n") {
			t = untab(strings.TrimRight(t, " "))
			body := strings.TrimLeft(t, " ")
			ls = append(ls, line{at + i, len(t) - len(body), body})
		}
	}
	n.Kids = parts(ls, g.id)
	spans(n.Kids)
	return n
}

// comment is a comment's text, its markers and the space after // gone.
func comment(c string) string {
	if t, ok := strings.CutPrefix(c, "//"); ok {
		return strings.TrimPrefix(t, " ")
	}
	return strings.TrimSuffix(strings.TrimPrefix(c, "/*"), "*/")
}

// text is the source from one place to another, as one line: its runs of
// space, line breaks too, a space each, and a brace that ends it gone. A
// declaration that does not parse may have no end, and then its line ends
// it.
func (g *golang) text(from, to token.Pos) string {
	a, b := g.fs.Position(from).Offset, g.fs.Position(to).Offset
	if !to.IsValid() || b < a {
		b = g.fs.Position(g.lineEnd(from)).Offset
	}
	s := strings.Join(strings.Fields(string(g.src[a:b])), " ")
	return strings.TrimSpace(strings.TrimSuffix(s, "{"))
}

// lineEnd is where the line pos is on ends.
func (g *golang) lineEnd(pos token.Pos) token.Pos {
	a := g.fs.Position(pos).Offset
	if i := bytes.IndexByte(g.src[a:], '\n'); i >= 0 {
		return pos + token.Pos(i)
	}
	return pos + token.Pos(len(g.src)-a)
}

// at is the address of the line pos is on, as the file numbers it.
func (g *golang) at(pos token.Pos) string {
	return At(g.id, g.fs.Position(pos).Line-g.off)
}

// opening is where a type's own line ends: at a struct's or an
// interface's brace, else at the type's end.
func opening(s *ast.TypeSpec) token.Pos {
	switch t := s.Type.(type) {
	case *ast.StructType:
		return t.Fields.Opening
	case *ast.InterfaceType:
		return t.Methods.Opening
	}
	return s.End()
}

// body is where a func's own line ends: at its body's brace, or at its
// end when it has none.
func body(d *ast.FuncDecl) token.Pos {
	if d.Body != nil {
		return d.Body.Lbrace
	}
	return d.End()
}

// recv is the name of the type a method is of, generic or not; "" for a
// func.
func recv(d *ast.FuncDecl) string {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return ""
	}
	t := d.Recv.List[0].Type
	if s, ok := t.(*ast.StarExpr); ok {
		t = s.X
	}
	switch i := t.(type) {
	case *ast.IndexExpr:
		t = i.X
	case *ast.IndexListExpr:
		t = i.X
	}
	if n, ok := t.(*ast.Ident); ok {
		return n.Name
	}
	return ""
}
