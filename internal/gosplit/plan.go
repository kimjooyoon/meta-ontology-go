package gosplit

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"path"
	"path/filepath"
)

func (planner Planner) PlanSource(root, subject string, limit int) (Plan, error) {
	var plan Plan
	if planner.normalize == nil {
		return plan, fmt.Errorf("import normalization adapter is required")
	}
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return plan, err
	}
	info, fset, file, source, err := parseSource(target, subject)
	if err != nil {
		return plan, err
	}
	sourceSHA := fmt.Sprintf("%x", sha256.Sum256(source))
	declarations := sourceDeclarations(file)
	positions := make(map[ast.Decl]int, len(declarations))
	for index, declaration := range declarations {
		positions[declaration] = index
	}
	groups, err := planner.partitionDeclarations(fset, file, declarations, limit)
	if err != nil {
		cause := fmt.Errorf("%w: %w", ErrBlocked, err)
		if _, ok := errors.AsType[*partitionUnavailable](err); ok {
			return plan, &Blocker{Kind: RenderedPlanUnavailable, Subject: subject, SourceSHA256: sourceSHA, cause: cause}
		}
		return plan, cause
	}
	if len(groups) < 2 {
		return plan, &Blocker{Kind: NoDeclarationSplit, Subject: subject, SourceSHA256: sourceSHA,
			cause: fmt.Errorf("%w: %s does not require declaration splitting", ErrBlocked, subject)}
	}
	plan = Plan{Directory: path.Dir(subject), Mode: info.Mode(), Parts: make([]Part, len(groups)), SourceSHA256: sourceSHA}
	for index, group := range groups {
		partSubject, nameErr := PartPath(subject, index+1)
		if nameErr != nil {
			return Plan{}, nameErr
		}
		data, renderErr := planner.RenderPart(fset, file, group)
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
