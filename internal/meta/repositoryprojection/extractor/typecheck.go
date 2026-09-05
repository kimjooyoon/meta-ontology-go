package extractor

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
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
	configuration := types.Config{Importer: newModuleImporter(root), Error: func(error) {}}
	checked, err := configuration.Check(filepath.ToSlash(filepath.Dir(logical)), fset, files, info)
	if err != nil && !sufficientFunctionTypeEvidence(function, info) {
		return typeEvidence{}, fail("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
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
	return typeEvidence{info: info, pkg: checked, files: files, funcs: functions}, nil
}

func sufficientFunctionTypeEvidence(function *ast.FuncDecl, info *types.Info) bool {
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
	sufficient := true
	ast.Inspect(function.Body, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || identifier.Name == "_" || packageSelectors[identifier] ||
			info.Defs[identifier] != nil || info.Uses[identifier] != nil {
			return true
		}
		sufficient = false
		return false
	})
	return sufficient
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
