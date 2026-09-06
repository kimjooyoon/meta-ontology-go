package extractor

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func checkTypes(root, logical string, fset *token.FileSet, file *ast.File, function *ast.FuncDecl) (typeEvidence, error) {
	files, err := packageTypeFiles(root, logical, fset, file)
	if err != nil {
		return typeEvidence{}, err
	}
	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Types:      map[ast.Expr]types.TypeAndValue{},
		Scopes:     map[ast.Node]*types.Scope{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	typeErrors := make([]error, 0)
	configuration := types.Config{Importer: newModuleImporter(root), Error: func(err error) {
		if err != nil {
			typeErrors = append(typeErrors, err)
		}
	}}
	checked, err := configuration.Check(filepath.ToSlash(filepath.Dir(logical)), fset, files, info)
	if err != nil {
		unresolved := unresolvedFunctionIdentifiers(function, info)
		if len(unresolved) != 0 {
			diagnostics := []string{
				"evidence=types-check",
				"go-types-error=" + normalizeTypeDiagnostic(root, err),
				"unresolved-identifiers=" + strings.Join(unresolved, ","),
			}
			if len(typeErrors) != 0 {
				errors := make([]string, 0, len(typeErrors))
				for _, typeError := range typeErrors {
					errors = append(errors, normalizeTypeDiagnostic(root, typeError))
				}
				sort.Strings(errors)
				diagnostics = append(diagnostics, "go-types-callback-errors="+strings.Join(errors, " || "))
			}
			return typeEvidence{}, failWithDiagnostics("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", diagnostics)
		}
	}
	functions := make(map[*types.Func]*ast.FuncDecl)
	for _, packageFile := range files {
		for _, declaration := range packageFile.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Name == nil {
				continue
			}
			if object, ok := info.Defs[function.Name].(*types.Func); ok {
				functions[object] = function
			}
		}
	}
	return typeEvidence{info: info, pkg: checked, files: files, funcs: functions, fset: fset}, nil
}

func sufficientFunctionTypeEvidence(function *ast.FuncDecl, info *types.Info) bool {
	return len(unresolvedFunctionIdentifiers(function, info)) == 0
}

func unresolvedFunctionIdentifiers(function *ast.FuncDecl, info *types.Info) []string {
	packageSelectors := make(map[*ast.Ident]bool)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		identifier, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, ok := info.Uses[identifier].(*types.PkgName); ok {
			packageSelectors[identifier] = true
		}
		return true
	})
	unresolved := make(map[string]bool)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || identifier.Name == "_" || packageSelectors[identifier] ||
			info.Defs[identifier] != nil || info.Uses[identifier] != nil {
			return true
		}
		unresolved[identifier.Name] = true
		return true
	})
	result := make([]string, 0, len(unresolved))
	for identifier := range unresolved {
		result = append(result, identifier)
	}
	sort.Strings(result)
	return result
}

func normalizeTypeDiagnostic(root string, err error) string {
	if err == nil {
		return "<nil>"
	}
	text := strings.TrimSpace(err.Error())
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\\", "/")
	text = strings.ReplaceAll(text, "\n", "\\n")
	workspace := strings.TrimRight(filepath.ToSlash(filepath.Clean(root)), "/")
	if workspace != "." && workspace != "" {
		text = strings.ReplaceAll(text, workspace+"/", "<workspace>/")
	}
	return text
}

func packageTypeFiles(root, logical string, fset *token.FileSet, target *ast.File) ([]*ast.File, error) {
	directory := filepath.Join(root, filepath.Dir(filepath.FromSlash(logical)))
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fail("derive-recipe", "load-test-package", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
	}
	testPackage := strings.HasSuffix(logical, "_test.go")
	targetName := filepath.Base(filepath.FromSlash(logical))
	files := []*ast.File{target}
	for _, entry := range entries {
		if entry.Name() == targetName || !isPackageTypeFile(entry.Name(), testPackage) || !buildFileMatches(directory, entry.Name()) {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			return nil, fail("derive-recipe", "load-test-package", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
		}
		parsed, parseErr := parser.ParseFile(fset, filepath.Join(directory, entry.Name()), data, parser.ParseComments)
		if parseErr != nil {
			return nil, fail("derive-recipe", "load-test-package", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
		}
		if parsed.Name.Name == target.Name.Name {
			files = append(files, parsed)
		}
	}
	return files, nil
}

func isPackageTypeFile(name string, includeTests bool) bool {
	if !strings.HasSuffix(name, ".go") {
		return false
	}
	return includeTests || !strings.HasSuffix(name, "_test.go")
}

func buildFileMatches(directory, name string) bool {
	matched, err := build.Default.MatchFile(directory, name)
	return err == nil && matched
}
