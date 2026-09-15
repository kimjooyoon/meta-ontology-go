package semanticbinding

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func collectRecords(
	parsed []parsedSource,
	fileSet *token.FileSet,
	info *types.Info,
	registered map[string]struct{},
) ([]Binding, []Obligation, error) {
	state := recordState{ids: make(map[string]Span), bindingTargets: make(map[string]Span)}
	for _, source := range parsed {
		attachments := attachmentsFor(source.file, source.input.PackagePath)
		for _, group := range source.file.Comments {
			for _, comment := range group.List {
				if err := state.addComment(comment, group, attachments, fileSet, info, registered); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return state.bindings, state.obligations, nil
}
func (state *recordState) addComment(
	comment *ast.Comment,
	group *ast.CommentGroup,
	attachments map[*ast.CommentGroup][]declarationTarget,
	fileSet *token.FileSet,
	info *types.Info,
	registered map[string]struct{},
) error {
	current, found, err := parseDirective(comment, spanFor(fileSet, comment))
	if err != nil || !found {
		return err
	}
	targets := attachments[group]
	if len(targets) == 0 {
		return bindingError(CodeDetachedComment, current.span, "directive is not attached to a declaration")
	}
	if len(targets) != 1 {
		return bindingError(CodeAmbiguousBinding, current.span, "directive is attached to multiple declarations")
	}
	current, err = validateDirective(current)
	if err != nil {
		return err
	}
	if err := ensureRegistered(current, registered); err != nil {
		return err
	}
	target := targets[0]
	key, err := targetObjectKey(target.node, info)
	if err != nil {
		return withErrorSpan(err, current.span)
	}
	declarationSpan := spanFor(fileSet, target.node)
	id := current.fields["id"]
	if previous, exists := state.ids[id]; exists {
		return bindingError(CodeDuplicateID, current.span, fmt.Sprintf("identity is already bound at %s", previous))
	}
	state.ids[id] = current.span
	return state.addRecord(current, target, key, declarationSpan)
}
