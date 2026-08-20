package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
)

func removeDeclaration(source parsedSource, declaration ast.Decl) ([]byte, []byte, error) {
	start := declaration.Pos()
	switch value := declaration.(type) {
	case *ast.FuncDecl:
		if value.Doc != nil {
			start = value.Doc.Pos()
		}
	case *ast.GenDecl:
		if value.Doc != nil {
			start = value.Doc.Pos()
		}
	}
	startOffset := source.Fset.Position(start).Offset
	endOffset := source.Fset.Position(declaration.End()).Offset
	if startOffset < 0 || endOffset < startOffset || endOffset > len(source.Source) {
		return nil, nil, fmt.Errorf("invalid declaration span in %s", source.Subject)
	}
	snippet := append([]byte(nil), bytes.TrimSpace(source.Source[startOffset:endOffset])...)
	for endOffset < len(source.Source) && isSpacing(source.Source[endOffset]) {
		endOffset++
	}
	remaining := make([]byte, 0, len(source.Source)-(endOffset-startOffset))
	remaining = append(remaining, source.Source[:startOffset]...)
	remaining = append(remaining, source.Source[endOffset:]...)
	formatted, err := formatPruned(remaining)
	if err != nil {
		return nil, nil, fmt.Errorf("format source remainder: %w", err)
	}
	return snippet, formatted, nil
}

func appendDeclaration(target parsedSource, snippet []byte, additions []importAddition) ([]byte, error) {
	packageEnd := target.Fset.Position(target.File.Name.End()).Offset
	if packageEnd < 0 || packageEnd > len(target.Source) {
		return nil, fmt.Errorf("invalid package span in %s", target.Subject)
	}
	combined := make([]byte, 0, len(target.Source)+len(snippet)+len(additions)*32)
	combined = append(combined, target.Source[:packageEnd]...)
	for _, addition := range additions {
		combined = append(combined, fmt.Sprintf("\n\nimport %s %q", addition.Name, addition.Path)...)
	}
	combined = append(combined, target.Source[packageEnd:]...)
	combined = bytes.TrimSpace(combined)
	combined = append(combined, '\n', '\n')
	combined = append(combined, snippet...)
	combined = append(combined, '\n')
	formatted, err := format.Source(combined)
	if err != nil {
		return nil, fmt.Errorf("format destination: %w", err)
	}
	return formatted, nil
}

func isSpacing(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}
