package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
)

var errSplitBlocked = errors.New("split blocked")

type splitPart struct {
	Path    string
	Subject string
	Data    []byte
}

type splitPlan struct {
	Directory string
	Mode      os.FileMode
	Parts     []splitPart
}

func planSource(root, subject string, limit int) (splitPlan, error) {
	var plan splitPlan
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return plan, err
	}
	source, err := os.ReadFile(target)
	if err != nil {
		return plan, fmt.Errorf("read %s: %w", subject, err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return plan, fmt.Errorf("stat %s: %w", subject, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, subject, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return plan, fmt.Errorf("parse %s: %w", subject, err)
	}
	declarations := sourceDeclarations(file)
	groups := make([][]ast.Decl, 0)
	current := make([]ast.Decl, 0)
	for _, declaration := range declarations {
		candidate := append(append([]ast.Decl(nil), current...), declaration)
		data, renderErr := renderPart(fset, file, candidate)
		if renderErr != nil {
			return plan, renderErr
		}
		if physicalLines(data) <= limit {
			current = candidate
			continue
		}
		if len(current) == 0 {
			position := fset.Position(declaration.Pos())
			return plan, fmt.Errorf("%w: declaration at %s:%d exceeds %d lines", errSplitBlocked, subject, position.Line, limit)
		}
		groups = append(groups, current)
		current = []ast.Decl{declaration}
		data, renderErr = renderPart(fset, file, current)
		if renderErr != nil || physicalLines(data) > limit {
			return plan, fmt.Errorf("%w: declaration after line %d exceeds %d lines", errSplitBlocked, positionLine(fset, declaration), limit)
		}
	}
	if len(current) != 0 {
		groups = append(groups, current)
	}
	if len(groups) < 2 {
		return plan, fmt.Errorf("%w: %s does not require declaration splitting", errSplitBlocked, subject)
	}
	plan = splitPlan{Directory: path.Dir(subject), Mode: info.Mode(), Parts: make([]splitPart, len(groups))}
	for index, group := range groups {
		partSubject, nameErr := splitPartPath(subject, index+1)
		if nameErr != nil {
			return splitPlan{}, nameErr
		}
		data, renderErr := renderPart(fset, file, group)
		if renderErr != nil {
			return splitPlan{}, renderErr
		}
		partPath := filepath.Join(filepath.Dir(target), filepath.Base(filepath.FromSlash(partSubject)))
		plan.Parts[index] = splitPart{Path: partPath, Subject: partSubject, Data: data}
	}
	return plan, nil
}

func sourceDeclarations(file *ast.File) []ast.Decl {
	declarations := make([]ast.Decl, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		general, isImport := declaration.(*ast.GenDecl)
		if !isImport || general.Tok != token.IMPORT {
			declarations = append(declarations, declaration)
		}
	}
	return declarations
}

func positionLine(fset *token.FileSet, declaration ast.Decl) int {
	return fset.Position(declaration.Pos()).Line
}
