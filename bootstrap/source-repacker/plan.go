package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func planRepack(root, subject string, limit int) (repackPlan, error) {
	source, err := loadSource(root, subject)
	if err != nil {
		return repackPlan{}, err
	}
	targets, err := destinationSources(source, limit)
	if err != nil {
		return repackPlan{}, err
	}
	declarations := movableDeclarations(source.File)
	for index := len(declarations) - 1; index >= 0; index-- {
		declaration := declarations[index]
		snippet, sourceAfter, transformErr := removeDeclaration(source, declaration)
		if transformErr != nil {
			return repackPlan{}, transformErr
		}
		if physicalLines(sourceAfter) > limit {
			continue
		}
		for _, target := range targets {
			if !targetSupports(source, target, declaration) {
				continue
			}
			targetAfter, appendErr := appendDeclaration(target.Source, snippet)
			if appendErr != nil {
				return repackPlan{}, appendErr
			}
			if physicalLines(targetAfter) > limit {
				continue
			}
			return repackPlan{Edits: []fileEdit{
				{Path: source.Path, Subject: source.Subject, Before: source.Source, After: sourceAfter, Mode: uint32(source.Mode)},
				{Path: target.Path, Subject: target.Subject, Before: target.Source, After: targetAfter, Mode: uint32(target.Mode)},
			}}, nil
		}
	}
	return repackPlan{}, fmt.Errorf("%w: %s has no safe destination", errRepackBlocked, subject)
}

func movableDeclarations(file *ast.File) []ast.Decl {
	result := make([]ast.Decl, 0)
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.FuncDecl:
			if value.Name.Name != "init" {
				result = append(result, value)
			}
		case *ast.GenDecl:
			if value.Tok == token.CONST || value.Tok == token.TYPE {
				result = append(result, value)
			}
		}
	}
	return result
}
