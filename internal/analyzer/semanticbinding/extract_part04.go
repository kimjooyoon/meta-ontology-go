package semanticbinding

import (
	"fmt"
	"go/ast"
)

func (state *recordState) addRecord(
	current directive,
	target declarationTarget,
	key string,
	declarationSpan Span,
) error {
	id := current.fields["id"]
	if current.kind == "bind" {
		targetID := target.inputKey(key)
		if previous, exists := state.bindingTargets[targetID]; exists {
			return bindingError(
				CodeAmbiguousBinding, current.span,
				fmt.Sprintf("declaration already has a binding at %s", previous),
			)
		}
		state.bindingTargets[targetID] = current.span
		state.bindings = append(state.bindings, Binding{
			ID: id, Role: Role(current.fields["role"]),
			PackagePath: target.packagePathValue(), DeclarationKey: key,
			Span: declarationSpan, DirectiveSpan: current.span,
		})
		return nil
	}
	state.obligations = append(state.obligations, Obligation{
		ID: id, Subject: current.fields["subject"], Pressure: current.fields["pressure"],
		PackagePath: target.packagePathValue(), DeclarationKey: key,
		Span: declarationSpan, DirectiveSpan: current.span,
	})
	return nil
}
func attachmentsFor(file *ast.File, packagePath string) map[*ast.CommentGroup][]declarationTarget {
	result := make(map[*ast.CommentGroup][]declarationTarget)
	for _, declaration := range file.Decls {
		switch current := declaration.(type) {
		case *ast.FuncDecl:
			if current.Doc != nil {
				result[current.Doc] = append(result[current.Doc], declarationTarget{
					node: current, file: file, packagePath: packagePath,
				})
			}
		case *ast.GenDecl:
			if current.Doc != nil {
				result[current.Doc] = append(result[current.Doc], declarationTarget{
					node: current, file: file, packagePath: packagePath,
				})
			}
			for _, specification := range current.Specs {
				if typeSpec, ok := specification.(*ast.TypeSpec); ok && typeSpec.Doc != nil {
					result[typeSpec.Doc] = append(result[typeSpec.Doc], declarationTarget{
						node: typeSpec, file: file, packagePath: packagePath,
					})
				}
			}
		}
	}
	return result
}
func (t declarationTarget) packagePathValue() string {
	return t.packagePath
}
func (t declarationTarget) inputKey(key string) string {
	return t.packagePath + "\x00" + key
}
