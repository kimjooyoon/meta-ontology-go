package syntaxregistration

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
)

type sourceEdit struct {
	start, end int
	text       string
}

type goSource struct {
	raw   []byte
	set   *token.FileSet
	file  *ast.File
	edits []sourceEdit
}

func parseGo(raw []byte) (*goSource, error) {
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "input.go", raw, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	return &goSource{raw: raw, set: set, file: file}, nil
}

func (source *goSource) text(node ast.Node) string {
	return string(source.raw[source.set.Position(node.Pos()).Offset:source.set.Position(node.End()).Offset])
}

func (source *goSource) replace(node ast.Node, text string) {
	source.edits = append(source.edits, sourceEdit{source.set.Position(node.Pos()).Offset, source.set.Position(node.End()).Offset, text})
}

func (source *goSource) insert(position token.Pos, text string) {
	offset := source.set.Position(position).Offset
	source.edits = append(source.edits, sourceEdit{offset, offset, text})
}

func (source *goSource) finish() ([]byte, error) {
	sort.Slice(source.edits, func(i, j int) bool { return source.edits[i].start > source.edits[j].start })
	out := bytes.Clone(source.raw)
	limit := len(out)
	for _, edit := range source.edits {
		if edit.start < 0 || edit.end < edit.start || edit.end > limit {
			return nil, fmt.Errorf("registration edits overlap or escape their source")
		}
		out = append(append(bytes.Clone(out[:edit.start]), []byte(edit.text)...), out[edit.end:]...)
		limit = edit.start
	}
	return format.Source(out)
}

func (source *goSource) function(name string) (*ast.FuncDecl, error) {
	var found *ast.FuncDecl
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name {
			if found != nil {
				return nil, fmt.Errorf("ambiguous native function %s", name)
			}
			found = function
		}
	}
	if found == nil {
		return nil, fmt.Errorf("native function missing: %s", name)
	}
	return found, nil
}

func integer(node ast.Expr) (int, bool) {
	literal, ok := node.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return 0, false
	}
	value, err := strconv.Atoi(literal.Value)
	return value, err == nil
}
