package syntaxregistration

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"slices"
	"strconv"
)

type sourceEdit struct {
	path     string
	start    int
	end      int
	text     string
	activity string
}

type goSource struct {
	set      *token.FileSet
	file     *ast.File
	units    map[string][]byte
	edits    []sourceEdit
	activity string
}

func (source *goSource) location(position token.Pos) (string, int) {
	file := source.set.File(position)
	return file.Name(), file.Offset(position)
}

func (source *goSource) text(node ast.Node) string {
	path, start := source.location(node.Pos())
	endPath, end := source.location(node.End())
	if path != endPath {
		return ""
	}
	return string(source.units[path][start:end])
}

func (source *goSource) replace(node ast.Node, text string) {
	path, start := source.location(node.Pos())
	_, end := source.location(node.End())
	source.edits = append(source.edits, sourceEdit{path, start, end, text, source.activity})
}

func (source *goSource) insert(position token.Pos, text string) {
	path, offset := source.location(position)
	source.edits = append(source.edits, sourceEdit{path, offset, offset, text, source.activity})
}

func (source *goSource) appendAt(anchor ast.Node, text string) {
	path, _ := source.location(anchor.Pos())
	end := len(source.units[path])
	source.edits = append(source.edits, sourceEdit{path, end, end, text, source.activity})
}

func (source *goSource) finish() (map[string][]byte, error) {
	grouped := map[string][]sourceEdit{}
	for _, edit := range source.edits {
		grouped[edit.path] = append(grouped[edit.path], edit)
	}
	out := map[string][]byte{}
	for _, path := range sortedPaths(grouped) {
		edits := grouped[path]
		slices.SortStableFunc(edits, func(a, b sourceEdit) int {
			if a.start > b.start {
				return -1
			}
			if a.start < b.start {
				return 1
			}
			return 0
		})
		raw := bytes.Clone(source.units[path])
		limit := len(raw)
		for _, edit := range edits {
			if edit.start < 0 || edit.end < edit.start || edit.end > limit || edit.activity == "" {
				return nil, fmt.Errorf("registration edits overlap, escape a source unit or lack an activity")
			}
			raw = append(append(bytes.Clone(raw[:edit.start]), []byte(edit.text)...), raw[edit.end:]...)
			limit = edit.start
		}
		formatted, err := format.Source(raw)
		if err != nil {
			return nil, err
		}
		out[path] = formatted
	}
	return out, nil
}

func (source *goSource) function(name string) (*ast.FuncDecl, error) {
	var found *ast.FuncDecl
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name != name {
			continue
		}
		if found != nil {
			return nil, failure("UNKNOWN", "bind-symbol:"+name, "REGISTRATION_SYMBOL_AMBIGUOUS",
				"AMBIGUOUS", "disambiguate-source-units")
		}
		found = function
	}
	if found == nil {
		return nil, failure("UNKNOWN", "bind-symbol:"+name, "REGISTRATION_SYMBOL_MISSING",
			"DIRECT_MISSING", "restore-source-unit")
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
