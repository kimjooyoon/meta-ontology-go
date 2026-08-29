package extractor

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

func checkTypes(root, logical string, fset *token.FileSet, file *ast.File) (typeEvidence, error) {
	files, err := packageTypeFiles(root, logical, fset, file)
	if err != nil {
		return typeEvidence{}, err
	}
	info := &types.Info{
		Defs:   map[*ast.Ident]types.Object{},
		Uses:   map[*ast.Ident]types.Object{},
		Scopes: map[ast.Node]*types.Scope{},
	}
	configuration := types.Config{Importer: importer.Default(), Error: func(error) {}}
	_, err = configuration.Check(filepath.ToSlash(filepath.Dir(logical)), fset, files, info)
	if err != nil {
		return typeEvidence{}, fail("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
	}
	return typeEvidence{info: info}, nil
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

func isPackageTypeFile(name string, testPackage bool) bool {
	if !strings.HasSuffix(name, ".go") {
		return false
	}
	return strings.HasSuffix(name, "_test.go") == testPackage
}

func buildFileMatches(directory, name string) bool {
	matched, err := build.Default.MatchFile(directory, name)
	return err == nil && matched
}
