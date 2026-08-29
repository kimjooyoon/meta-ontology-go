package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
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
	if groups, ok := orderedDeclarationPartition(fset, file, declarations, limit); ok {
		return groups, nil
	}
	if groups, ok := orderedGreedyDeclarationPartition(fset, file, declarations, limit); ok {
		return groups, nil
	}
	return nil, fmt.Errorf("declarations cannot preserve source order within %d lines", limit)
}

func declarationIdentity(fset *token.FileSet, declaration ast.Decl, _ int) (string, error) {
	var output bytes.Buffer
	if err := format.Node(&output, fset, declaration); err != nil {
		return "", err
	}
	sum := sha256.Sum256(output.Bytes())
	return fmt.Sprintf("%x", sum), nil
}

func orderedGreedyDeclarationPartition(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, bool) {
	groups, err := greedyDeclarationPartition(fset, file, declarations, limit)
	return groups, err == nil && len(groups) > 1
}

func orderedDeclarationPartition(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, bool) {
	for start := 1; start < len(declarations); start++ {
		prefix := append([]ast.Decl{}, declarations[:start]...)
		suffix := append([]ast.Decl{}, declarations[start:]...)
		if !allMovableDeclarations(suffix) {
			continue
		}
		prefixData, prefixErr := renderPart(fset, file, prefix)
		suffixData, suffixErr := renderPart(fset, file, suffix)
		if prefixErr == nil && suffixErr == nil && physicalLines(prefixData) <= limit && physicalLines(suffixData) <= limit {
			return [][]ast.Decl{prefix, suffix}, true
		}
	}
	return nil, false
}

func middleDeclarationPartition(fset *token.FileSet, file *ast.File, declarations []ast.Decl, limit int) ([][]ast.Decl, bool) {
	if !allMovableDeclarations(declarations) {
		return nil, false
	}
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
