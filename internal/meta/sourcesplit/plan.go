package sourcesplit

import (
	"fmt"
	"go/ast"
	"path"
	"path/filepath"
)

func PlanSource(root, subject string, limit int) (Plan, error) {
	var plan Plan
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return plan, err
	}
	info, fset, file, err := parseSource(target, subject)
	if err != nil {
		return plan, err
	}
	declarations := sourceDeclarations(file)
	positions := make(map[ast.Decl]int, len(declarations))
	for index, declaration := range declarations {
		positions[declaration] = index
	}
	groups, err := partitionDeclarations(fset, file, declarations, limit)
	if err != nil {
		return plan, &Blocker{"RENDERED_DECLARATION_CAPACITY", err.Error()}
	}
	if len(groups) < 2 {
		return plan, &Blocker{"SPLIT_NOT_REQUIRED", fmt.Sprintf("%s does not require declaration splitting", subject)}
	}
	plan = Plan{Directory: path.Dir(subject), Mode: info.Mode(), Parts: make([]Part, len(groups))}
	for index, group := range groups {
		partSubject, nameErr := splitPartPath(subject, index+1)
		if nameErr != nil {
			return Plan{}, nameErr
		}
		data, renderErr := renderPart(fset, file, group)
		if renderErr != nil {
			return Plan{}, renderErr
		}
		partPath := filepath.Join(filepath.Dir(target), filepath.Base(filepath.FromSlash(partSubject)))
		order := make([]DeclarationOrder, 0, len(group))
		for _, declaration := range group {
			identity, identityErr := declarationIdentity(fset, declaration, positions[declaration])
			if identityErr != nil {
				return Plan{}, identityErr
			}
			order = append(order, identity)
		}
		plan.Parts[index] = Part{Path: partPath, Subject: partSubject, Data: data, DeclarationOrder: order}
	}
	return plan, nil
}
