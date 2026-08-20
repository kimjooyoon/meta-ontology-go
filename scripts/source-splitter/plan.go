package main

import (
	"fmt"
	"go/ast"
	"path"
	"path/filepath"
)

func planSource(root, subject string, limit int) (splitPlan, error) {
	var plan splitPlan
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return plan, err
	}
	info, fset, file, err := parseSource(target, subject)
	if err != nil {
		return plan, err
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
