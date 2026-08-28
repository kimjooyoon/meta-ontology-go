package main

import (
	"fmt"
	"go/ast"
)

type extractionFailure struct {
	Stage, Step, Reason, UnknownClass, NextOperation string
	BlockedBy                                     []string
}

func (e extractionFailure) Error() string {
	return fmt.Sprintf("%s/%s/%s unknown_class=%s next=%s blocked_by=%v", e.Stage, e.Step, e.Reason, e.UnknownClass, e.NextOperation, e.BlockedBy)
}

func extractionError(stage, step, reason, class, next string, blocked []string) error {
	return extractionFailure{stage, step, reason, class, next, append([]string(nil), blocked...)}
}

type astDecl struct {
	decl ast.Decl
	start, end, order int
	identity           string
}

type astImport struct {
	decl *ast.GenDecl
	spec *ast.ImportSpec
	path, name string
}

type astRendered struct {
	source, helper []byte
}
