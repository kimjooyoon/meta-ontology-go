package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func expressionCandidates(name string, data []byte) ([]sourceSpan, error) {
	files := token.NewFileSet()
	source, err := parser.ParseFile(files, name, data,
		parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	spans := make([]sourceSpan, 0)
	ast.Inspect(source, func(node ast.Node) bool {
		if isDensityExpression(node) && multilineNode(files, node) && commentFree(source, node) {
			spans = append(spans, sourceSpan{
				start: files.Position(node.Pos()).Offset,
				end:   files.Position(node.End()).Offset,
			})
		}
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for index := 0; index+1 < len(block.List); index++ {
			first, second := block.List[index], block.List[index+1]
			if simpleStatement(first) && simpleStatement(second) &&
				files.Position(first.End()).Line < files.Position(second.Pos()).Line &&
				commentFree(source, &statementPair{first: first, second: second}) {
				spans = append(spans, sourceSpan{
					start: files.Position(first.Pos()).Offset,
					end:   files.Position(second.End()).Offset,
				})
			}
		}
		return true
	})
	return spans, nil
}

func isDensityExpression(node ast.Node) bool {
	switch node.(type) {
	case *ast.BinaryExpr, *ast.CallExpr, *ast.CompositeLit:
		return true
	default:
		return false
	}
}

func multilineNode(files *token.FileSet, node ast.Node) bool {
	return files.Position(node.Pos()).Line < files.Position(node.End()).Line
}

func commentFree(source *ast.File, node ast.Node) bool {
	for _, comment := range source.Comments {
		if comment.Pos() >= node.Pos() && comment.End() <= node.End() {
			return false
		}
	}
	return true
}
