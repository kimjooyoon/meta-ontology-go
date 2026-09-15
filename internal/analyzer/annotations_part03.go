package analyzer

import (
	"go/ast"
	"go/token"
)

func registrationFor(file parsedFile, fileSet *token.FileSet, parsed annotation, defaultKind SymbolKind, ref SymbolRef) (Registration, bool, Diagnostics) {
	if !parsed.active {
		return Registration{}, false, nil
	}
	span := declarationSpan(fileSet, ref, file.file)
	if parsed.conflict {
		return Registration{}, false, Diagnostics{{
			Code:    DiagConflictingAnnotation,
			Message: "semantic annotation contains conflicting values",
			Span:    span,
		}}
	}
	if parsed.id == "" {
		return Registration{}, false, Diagnostics{{
			Code:    DiagInvalidAnnotation,
			Message: "semantic annotation requires an id",
			Span:    span,
		}}
	}
	if parsed.kind == "" {
		parsed.kind = defaultKind
	}
	registration := Registration{
		Ref:      ref,
		Kind:     parsed.kind,
		Identity: Identity{Namespace: parsed.namespace, ID: parsed.id},
		Span:     spanFor(fileSet, file.file),
	}
	if ref.Receiver != "" || ref.Name != "" {
		registration.Span = span
	}
	if err := validateRegistration(registration); err != nil {
		return Registration{}, false, Diagnostics{{
			Code:    DiagInvalidAnnotation,
			Message: err.Error(),
			Span:    span,
		}}
	}
	return registration, true, nil
}
func declarationSpan(fileSet *token.FileSet, ref SymbolRef, file *ast.File) Span {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == ref.Name && receiverTypeName(function.Recv) == ref.Receiver {
			return spanFor(fileSet, function)
		}
		if group, ok := declaration.(*ast.GenDecl); ok {
			for _, specification := range group.Specs {
				if namedSpec(specification, ref.Name) {
					return spanFor(fileSet, specification)
				}
			}
		}
	}
	return spanFor(fileSet, file)
}
