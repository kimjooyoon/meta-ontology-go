package semanticbinding

import (
	"go/ast"
	"go/token"
)

func spanFor(fileSet *token.FileSet, node ast.Node) Span {
	if node == nil {
		return Span{}
	}
	start := fileSet.PositionFor(node.Pos(), true)
	end := fileSet.PositionFor(node.End(), true)
	return Span{
		Filename: start.Filename,
		Start:    Position{Offset: start.Offset, Line: start.Line, Column: start.Column},
		End:      Position{Offset: end.Offset, Line: end.Line, Column: end.Column},
	}
}
