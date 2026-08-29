package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func partitionDeclarations(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, error) {
	all, err := renderPart(fset, file, declarations)
	if err != nil {
		return nil, err
	}
	if physicalLines(all) <= limit {
		return [][]ast.Decl{declarations}, nil
	}
	if groups, ok := middleDeclarationPartition(fset, file, declarations, limit); ok {
		return groups, nil
	}
	return greedyDeclarationPartition(fset, file, declarations, limit)
}

func middleDeclarationPartition(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, bool) {
	for start := 0; start < len(declarations)-1; start++ {
		for end := start + 1; end < len(declarations); end++ {
			if !movableDeclaration(declarations[end-1]) {
				break
			}
			if !allMovableDeclarations(declarations[start:end]) {
				continue
			}
			remaining := appendDeclarationRange(declarations[:start], declarations[end:])
			moved := append([]ast.Decl{}, declarations[start:end]...)
			remainingData, remainingErr := renderPart(fset, file, remaining)
			movedData, movedErr := renderPart(fset, file, moved)
			if remainingErr == nil && movedErr == nil && physicalLines(remainingData) <= limit && physicalLines(movedData) <= limit {
				return [][]ast.Decl{remaining, moved}, true
			}
		}
	}
	return nil, false
}

func greedyDeclarationPartition(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, error) {
	groups := make([][]ast.Decl, 0)
	current := make([]ast.Decl, 0)
	for _, declaration := range declarations {
		candidate := append(append([]ast.Decl{}, current...), declaration)
		data, err := renderPart(fset, file, candidate)
		if err != nil {
			return nil, err
		}
		if physicalLines(data) <= limit {
			current = candidate
			continue
		}
		if len(current) == 0 {
			return nil, fmt.Errorf("declaration at line %d exceeds %d lines", positionLine(fset, declaration), limit)
		}
		groups = append(groups, current)
		current = []ast.Decl{declaration}
		data, err = renderPart(fset, file, current)
		if err != nil || physicalLines(data) > limit {
			return nil, fmt.Errorf("declaration after line %d exceeds %d lines", positionLine(fset, declaration), limit)
		}
	}
	if len(current) != 0 {
		groups = append(groups, current)
	}
	return groups, nil
}

func movableDeclaration(declaration ast.Decl) bool {
	function, ok := declaration.(*ast.FuncDecl)
	if ok {
		return function.Name != nil && function.Name.Name != "init"
	}
	general, ok := declaration.(*ast.GenDecl)
	return ok && general.Tok == token.TYPE
}

func allMovableDeclarations(declarations []ast.Decl) bool {
	for _, declaration := range declarations {
		if !movableDeclaration(declaration) {
			return false
		}
	}
	return true
}

func appendDeclarationRange(first, second []ast.Decl) []ast.Decl {
	result := make([]ast.Decl, 0, len(first)+len(second))
	result = append(result, first...)
	result = append(result, second...)
	return result
}
