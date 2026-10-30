package main

import (
	"go/ast"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

// Keep the existing CLI tests and their independent expected values unchanged.
func physicalLines(data []byte) int { return gosplit.PhysicalLines(data) }

func splitPartPath(subject string, index int) (string, error) {
	return gosplit.PartPath(subject, index)
}

func renderPart(fset *token.FileSet, file *ast.File, declarations []ast.Decl) ([]byte, error) {
	return generation.NewGoSplitPlanner().RenderPart(fset, file, declarations)
}
