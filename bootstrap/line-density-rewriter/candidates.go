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
		declaration, isImport := node.(*ast.GenDecl)
		if isImport && declaration.Tok == token.IMPORT && declaration.Lparen.IsValid() && len(declaration.Specs) == 1 {
			specification := declaration.Specs[0]
			start := files.Position(specification.Pos()).Offset
			end := files.Position(specification.End()).Offset
			if inline, ok := oneLineTokens(data[start:end]); ok {
				candidates = append(candidates, sourceSpan{
					start: files.Position(declaration.Pos()).Offset,
					end: files.Position(declaration.End()).Offset,
					replacement: "import " + inline,
				})
			}
		}
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

func simpleStatement(statement ast.Stmt) bool {
	switch statement.(type) {
	case *ast.AssignStmt, *ast.ExprStmt, *ast.IncDecStmt, *ast.SendStmt:
		return true
	default:
		return false
	}
}

type statementPair struct {
	first  ast.Stmt
	second ast.Stmt
}

func (pair *statementPair) Pos() token.Pos { return pair.first.Pos() }
func (pair *statementPair) End() token.Pos { return pair.second.End() }
