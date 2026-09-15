package analyzer

import (
	"go/ast"
	"go/token"
)

type resolver struct {
	registry []Registration
	locals   []Registration
	fileSet  *token.FileSet
	imports  map[*ast.File]importTable
}

func newResolver(registry *Registry, fileSet *token.FileSet, files []parsedFile) *resolver {
	resolver := &resolver{
		registry: registry.all(),
		fileSet:  fileSet,
		imports:  make(map[*ast.File]importTable, len(files)),
	}
	for _, file := range files {
		resolver.imports[file.file] = importsFor(file.file)
	}
	return resolver
}
func (r *resolver) addLocal(registration Registration) bool {
	if err := validateRegistration(registration); err != nil {
		return false
	}
	for _, existing := range r.locals {
		if sameRegistration(existing, registration) {
			return false
		}
	}
	r.locals = append(r.locals, registration)
	return true
}
func (r *resolver) analyzeFile(file parsedFile, delta *SemanticDelta) {
	for _, declaration := range file.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		activity, ok := r.activityFor(function, file)
		if !ok {
			continue
		}
		collector := newFactCollector(r, delta, activity.Identity, file, function)
		collector.collectSignature(function)
		ast.Walk(collector, function.Body)
	}
}
func (r *resolver) activityFor(function *ast.FuncDecl, file parsedFile) (Registration, bool) {
	ref := SymbolRef{
		PackagePath: file.packagePath,
		PackageName: file.packageName,
		Receiver:    receiverTypeName(function.Recv),
		Name:        function.Name.Name,
	}
	result := r.resolve(ref)
	if result.state != resolved || len(result.entries) != 1 || result.entries[0].Kind != KindActivity {
		return Registration{}, false
	}
	return result.entries[0], true
}

type resolutionState uint8

const (
	unresolved resolutionState = iota
	resolved
	ambiguous
)
