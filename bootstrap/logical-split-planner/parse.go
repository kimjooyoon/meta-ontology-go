package main

import (
	"go/ast"
	"go/parser"
	"go/token"
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
			atom.movable = typed.Tok == token.CONST || typed.Tok == token.TYPE
			if typed.Doc != nil {
				start = typed.Doc.Pos()
			}
		case *ast.FuncDecl:
			atom.kind = "func"
			atom.movable = typed.Name.Name != "init"
			if typed.Doc != nil {
				start = typed.Doc.Pos()
			}
		default:
			atom.kind = "unknown"
		}
		atom.lines = files.Position(end).Line - files.Position(start).Line + 1
		atoms = append(atoms, atom)
	}
	return atoms, nil
}
