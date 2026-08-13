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

type resolution struct {
	state   resolutionState
	entries []Registration
}

func (r *resolver) resolve(ref SymbolRef) resolution {
	return makeResolution(r.lookup(ref))
}

func (r *resolver) resolveExpression(expr ast.Expr, file parsedFile, varTypes map[string]typeReference) resolution {
	switch expression := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return r.resolve(SymbolRef{
			PackagePath: file.packagePath,
			PackageName: file.packageName,
			Name:        expression.Name,
		})
	case *ast.SelectorExpr:
		return r.resolveSelector(expression, file, varTypes)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return r.resolveExpression(indexBase(expr), file, varTypes)
	default:
		return resolution{state: unresolved}
	}
}

func (r *resolver) resolveSelector(selector *ast.SelectorExpr, file parsedFile, varTypes map[string]typeReference) resolution {
	base := unwrapExpr(selector.X)
	if ident, ok := base.(*ast.Ident); ok {
		if typeRef, typed := varTypes[ident.Name]; typed {
			return r.resolve(SymbolRef{
				PackagePath: typeRef.packagePath,
				PackageName: typeRef.packageName,
				Receiver:    typeRef.name,
				Name:        selector.Sel.Name,
			})
		}
		if path, imported := r.imports[file.file].aliases[ident.Name]; imported {
			return r.resolve(SymbolRef{PackagePath: path, PackageName: ident.Name, Name: selector.Sel.Name})
		}
	}
	return resolution{state: unresolved}
}

func (r *resolver) lookup(ref SymbolRef) []Registration {
	if ref.Name == "" {
		return nil
	}
	all := r.allRegistrations()
	var exact []Registration
	for _, entry := range all {
		if !sameSymbol(entry.Ref, ref) || ref.PackagePath == "" {
			continue
		}
		if entry.Ref.PackagePath == ref.PackagePath {
			exact = append(exact, entry)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	var packageName []Registration
	for _, entry := range all {
		if !sameSymbol(entry.Ref, ref) || ref.PackageName == "" {
			continue
		}
		if entry.Ref.PackageName == ref.PackageName && (entry.Ref.PackagePath == "" || ref.PackagePath == "") {
			packageName = append(packageName, entry)
		}
	}
	return packageName
}

func sameSymbol(left, right SymbolRef) bool {
	return left.Name == right.Name && left.Receiver == right.Receiver
}

func (r *resolver) allRegistrations() []Registration {
	all := make([]Registration, 0, len(r.registry)+len(r.locals))
	all = append(all, r.registry...)
	all = append(all, r.locals...)
	return uniqueRegistrations(all)
}

func makeResolution(entries []Registration) resolution {
	switch len(entries) {
	case 0:
		return resolution{state: unresolved}
	case 1:
		return resolution{state: resolved, entries: entries}
	default:
		return resolution{state: ambiguous, entries: entries}
	}
}
