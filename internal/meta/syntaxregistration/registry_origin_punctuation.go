package syntaxregistration

import (
	"go/ast"
	"go/scanner"
	"go/token"
)

// Inline literals may omit a trailing comma. Scan tokens rather than comments,
// whose prose may itself contain commas, before inserting a new source member.
func (source *goSource) registryTailNeedsComma(list *ast.CompositeLit) bool {
	last := list.Elts[len(list.Elts)-1]
	path, start := source.location(last.End())
	_, end := source.location(list.Rbrace)
	raw := source.units[path][start:end]
	set := token.NewFileSet()
	file := set.AddFile(path, -1, len(raw))
	var lexer scanner.Scanner
	lexer.Init(file, raw, nil, 0)
	_, kind, _ := lexer.Scan()
	return kind != token.COMMA
}
