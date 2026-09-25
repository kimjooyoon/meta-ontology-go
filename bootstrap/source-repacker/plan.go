package main

import (
	"go/ast"
	"go/token"
	"slices"
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
	reducible, supported := false, false
	for _, declaration := range slices.Backward(declarations) {

		snippet, sourceAfter, transformErr := removeDeclaration(source, declaration)
		if transformErr != nil {
			return repackPlan{}, transformErr
		}
		if physicalLines(sourceAfter) > limit {
			continue
		}
		reducible = true
		for _, target := range targets {
			additions, compatible := requiredImports(source, target, declaration)
			if !compatible {
				continue
			}
			supported = true
			targetAfter, appendErr := appendDeclaration(target, snippet, additions)
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
	reason := blockedReason(len(declarations), len(targets), reducible, supported)
	return repackPlan{}, repackBlocked{Subject: subject, Reason: reason}
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
