package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
)

func declarationAtoms(name string, data []byte) ([]declarationAtom, error) {
	files := token.NewFileSet()
	source, err := parser.ParseFile(files, name, data,
		parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	atoms := make([]declarationAtom, 0, len(source.Decls))
	for _, declaration := range source.Decls {
		start, end := declaration.Pos(), declaration.End()
		atom := declarationAtom{}
		switch typed := declaration.(type) {
		case *ast.GenDecl:
			if typed.Tok == token.IMPORT {
				continue
			}
			atom.kind = typed.Tok.String()
			atom.identity = declarationIdentity(files, typed)
			atom.movable = typed.Tok == token.CONST || typed.Tok == token.TYPE
			atom.compactable = typed.Tok == token.VAR && hasDensityCarrier(typed)
			if typed.Doc != nil {
				start = typed.Doc.Pos()
			}
		case *ast.FuncDecl:
			atom.kind = "func"
			atom.identity = declarationIdentity(files, typed)
			atom.movable = typed.Name.Name != "init"
			atom.compactable = hasDensityCarrier(typed)
			if typed.Doc != nil {
				start = typed.Doc.Pos()
			}
		default:
			atom.kind = "unknown"
			atom.identity = atom.kind + "@" + strconv.Itoa(files.Position(declaration.Pos()).Offset)
		}
		atom.lines = files.Position(end).Line - files.Position(start).Line + 1
		atoms = append(atoms, atom)
	}
	return atoms, nil
}

func declarationIdentity(files *token.FileSet, declaration ast.Decl) string {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		if typed.Recv == nil {
			return "func:" + typed.Name.Name
		}
		var receiver bytes.Buffer
		if err := format.Node(&receiver, files, typed.Recv); err == nil {
			return "method:" + receiver.String() + ":" + typed.Name.Name
		}
		return "method-at:" + strconv.Itoa(files.Position(typed.Pos()).Offset) + ":" + typed.Name.Name
	case *ast.GenDecl:
		for _, raw := range typed.Specs {
			var name string
			switch spec := raw.(type) {
			case *ast.TypeSpec:
				name = spec.Name.Name
			case *ast.ValueSpec:
				if len(spec.Names) > 0 {
					name = spec.Names[0].Name
				}
			}
			if name != "" {
				return typed.Tok.String() + ":" + name
			}
		}
		return typed.Tok.String() + "@" + strconv.Itoa(files.Position(typed.Pos()).Offset)
	default:
		return "declaration@" + strconv.Itoa(files.Position(declaration.Pos()).Offset)
	}
}

func hasDensityCarrier(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(current ast.Node) bool {
		switch current.(type) {
		case *ast.CompositeLit, *ast.FuncLit:
			found = true
			return false
		}
		return !found
	})
	return found
}
