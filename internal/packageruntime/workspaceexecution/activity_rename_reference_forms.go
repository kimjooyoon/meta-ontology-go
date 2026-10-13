package workspaceexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func (index workspaceRenameIndex) assemblyTokenReferences(key string, d workspaceCallDeclaration,
	previous, token syntax.Token, span syntax.Span, choice int) ([]renameReference, error) {
	switch {
	case previous.Kind == syntax.TokenActivity || previous.Kind == syntax.TokenIdentifier && previous.Value == "oracle":
		return index.namedReference(key, d, token.Value, span, "assembly_activity")
	case previous.Kind == syntax.TokenIdentifier && previous.Value == "baseline":
		return index.literalReferences(key, d, span, token.Value, "baseline", true, false)
	case previous.Kind == syntax.TokenIdentifier && previous.Value == "alternative":
		if choice < 0 || choice >= len(d.activity.Assembly.Spec.Choices) {
			return nil, fmt.Errorf("alternative has no source choice")
		}
		selected := d.activity.Assembly.Spec.Choices[choice]
		if selected.Kind == "field_value" || selected.Kind == "field_update" {
			return index.literalReferences(key, d, span, token.Value, "choice:"+selected.ID, false, true)
		}
	}
	return nil, nil
}

func (index workspaceRenameIndex) namedReference(key string, d workspaceCallDeclaration,
	value string, span syntax.Span, surface string) ([]renameReference, error) {
	expression, err := parser.ParseExpr(value + "()")
	if err != nil {
		return nil, err
	}
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return nil, fmt.Errorf("named activity reference is unresolved")
	}
	callee, err := index.calls.resolveCall(d, call)
	if err != nil {
		return nil, err
	}
	start := strings.LastIndexByte(value, '.') + 1
	raw := index.sources[d.ref.PackagePath][d.source][span.Start.Offset:span.End.Offset]
	first, last, err := bodycodegen.SourceLiteralByteRange(raw, value, start, len(value))
	if err != nil {
		return nil, err
	}
	return []renameReference{{d.ref.PackagePath, d.source, index.identities[key], index.identities[callee], index.calls[callee].ref.PackagePath, surface,
		span.Start.Offset + first, span.Start.Offset + last, start != 0, true}}, nil
}

func (index workspaceRenameIndex) bindingReferences() ([]renameReference, error) {
	var refs []renameReference
	for pkg, files := range index.files {
		for _, file := range files {
			aliases := importAliases(pkg, file.Imports)
			for _, binding := range file.Bindings {
				for side, endpoint := range []syntax.BindingEndpoint{binding.Producer, binding.Consumer} {
					owner := pkg
					if endpoint.PackageAlias != "" {
						owner = aliases[endpoint.PackageAlias]
					}
					id := index.identities[packageActivityKey(owner, endpoint.Activity.Name)]
					if id == "" {
						return nil, fmt.Errorf("binding activity %s.%s has no resolved identity", owner, endpoint.Activity.Name)
					}
					span := endpoint.Activity.Span
					refs = append(refs, renameReference{pkg, span.Filename, "", id, owner, fmt.Sprintf("binding:%d", side),
						span.Start.Offset, span.End.Offset, endpoint.PackageAlias != "", false})
				}
			}
		}
	}
	return refs, nil
}
