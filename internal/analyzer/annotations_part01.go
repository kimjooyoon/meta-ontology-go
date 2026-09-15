package analyzer

import (
	"go/ast"
	"go/token"
)

type annotation struct {
	kind      SymbolKind
	namespace string
	id        string
	active    bool
	conflict  bool
}

func collectRegistrations(file parsedFile, fileSet *token.FileSet) ([]Registration, Diagnostics) {
	var registrations []Registration
	var diagnostics Diagnostics
	for _, declaration := range file.file.Decls {
		switch current := declaration.(type) {
		case *ast.FuncDecl:
			parsed := parseAnnotations(current.Doc)
			registration, ok, found := registrationFor(file, fileSet, parsed, KindActivity, SymbolRef{
				PackagePath: file.packagePath,
				PackageName: file.packageName,
				Receiver:    receiverTypeName(current.Recv),
				Name:        current.Name.Name,
			})
			diagnostics = append(diagnostics, found...)
			if ok {
				registrations = append(registrations, registration)
			}
		case *ast.GenDecl:
			foundRegistrations, foundDiagnostics := collectGenRegistrations(file, fileSet, current)
			registrations = append(registrations, foundRegistrations...)
			diagnostics = append(diagnostics, foundDiagnostics...)
		}
	}
	return registrations, diagnostics
}
