package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func guardCandidates(name string, data []byte) ([]sourceSpan, error) {
	files := token.NewFileSet()
	source, err := parser.ParseFile(files, name, data,
		parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	candidates := make([]sourceSpan, 0)
	ast.Inspect(source, func(node ast.Node) bool {
		statement, ok := node.(*ast.IfStmt)
		if !ok || statement.Else != nil || len(statement.Body.List) != 1 {
			return true
		}
		for _, comment := range source.Comments {
			if comment.Pos() >= statement.Pos() && comment.End() <= statement.End() {
				return true
			}
		}
		start := files.Position(statement.Pos()).Offset
		end := files.Position(statement.End()).Offset
		if files.Position(statement.End()).Line > files.Position(statement.Pos()).Line {
			candidates = append(candidates, sourceSpan{start: start, end: end})
		}
		return true
	})
	return candidates, nil
}
