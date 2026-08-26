package analyzer

import (
	"go/ast"
	"go/token"
)

func collectGenRegistrations(file parsedFile, fileSet *token.FileSet, declaration *ast.GenDecl) ([]Registration, Diagnostics) {
	var registrations []Registration
	var diagnostics Diagnostics
	for _, specification := range declaration.Specs {
		switch current := specification.(type) {
		case *ast.TypeSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			registration, ok, found := registrationFor(file, fileSet, parsed, KindEntity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Name:        current.Name.Name,
			})
			diagnostics = append(diagnostics, found...)
			if ok {
				if registration.Kind == "" {
					registration.Kind = KindEntity
				}
				registrations = append(registrations, registration)
			}
		case *ast.ValueSpec:
			parsed := mergeAnnotations(parseAnnotations(declaration.Doc), parseAnnotations(current.Doc))
			if !parsed.active {
				continue
			}
			for _, name := range current.Names {
				registration, ok, found := registrationFor(file, fileSet, parsed, "", SymbolRef{
					PackagePath: file.packagePath,
					PackageName: file.packageName,
					Name:        name.Name,
				})
				diagnostics = append(diagnostics, found...)
				if ok {
					registrations = append(registrations, registration)
				}
			}
		}
	}
	return registrations, diagnostics
}
