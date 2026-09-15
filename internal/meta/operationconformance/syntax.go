package operationconformance

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
)

func parseEvidence(file FileEvidence) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file.Path, file.Data, parser.ParseComments|parser.AllErrors)
	return fset, parsed, err
}

func sortedCandidates(files []FileEvidence) []FileEvidence {
	result := append([]FileEvidence(nil), files...)
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
