package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

func genericCandidates(fset *token.FileSet, file *ast.File) ([]astDecl, error) {
	seen := map[string]bool{}
	var out []astDecl
	for order, decl := range file.Decls {
		identity, movable := genericIdentity(fset, decl)
		if !movable {
			continue
		}
		if seen[identity] {
			return nil, extractionError("validate-ast", "identity", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{identity})
		}
		seen[identity] = true
		start := decl.Pos()
		if documented, ok := decl.(*ast.FuncDecl); ok && documented.Doc != nil {
			start = documented.Doc.Pos()
		} else if documented, ok := decl.(*ast.GenDecl); ok && documented.Doc != nil {
			start = documented.Doc.Pos()
		}
		out = append(out, astDecl{decl: decl, start: fset.Position(start).Offset, end: fset.Position(decl.End()).Offset, order: order, identity: identity})
	}
	return out, nil
}

func genericIdentity(fset *token.FileSet, decl ast.Decl) (string, bool) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" || d.Body == nil {
			return "", false
		}
		if d.Recv == nil {
			return "func:" + d.Name.Name, true
		}
		var receiver bytes.Buffer
		if format.Node(&receiver, fset, d.Recv) != nil {
			return "", false
		}
		return "method:" + receiver.String() + ":" + d.Name.Name, true
	case *ast.GenDecl:
		if d.Tok != token.CONST && d.Tok != token.TYPE {
			return "", false
		}
		for _, spec := range d.Specs {
			name := ""
			switch s := spec.(type) {
			case *ast.TypeSpec:
				name = s.Name.Name
			case *ast.ValueSpec:
				if len(s.Names) > 0 {
					name = s.Names[0].Name
				}
			}
			if name != "" {
				return d.Tok.String() + ":" + name, true
			}
		}
	}
	return "", false
}
