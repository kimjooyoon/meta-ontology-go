package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roleDeclarationOwner39247 struct {
	logical  string
	raw      []byte
	set      *token.FileSet
	file     *ast.File
	function *ast.FuncDecl
}

func roleDeclarationOwners39247(root string, required map[string]bool) ([]roleDeclarationOwner39247, error) {
	logicalDir := "internal/meta/syntaxregistration"
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(logicalDir)))
	if err != nil {
		return nil, err
	}
	var owners []roleDeclarationOwner39247
	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		found, err := roleDeclarationFile39247(root, logicalDir+"/"+entry.Name(), required)
		if err != nil {
			return nil, err
		}
		for _, owner := range found {
			name := owner.function.Name.Name
			if seen[name] {
				return nil, fmt.Errorf("role declaration has repeated owners: %s", name)
			}
			seen[name] = true
			owners = append(owners, owner)
		}
	}
	if len(required) == 0 || len(seen) != len(required) {
		return nil, fmt.Errorf("role declarations omitted: found %d of %d", len(seen), len(required))
	}
	return owners, nil
}

func roleDeclarationFile39247(root, logical string, required map[string]bool) ([]roleDeclarationOwner39247, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
	if err != nil {
		return nil, err
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, logical, raw, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var owners []roleDeclarationOwner39247
	if file.Name.Name == "syntaxregistration" {
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && required[function.Name.Name] {
				owners = append(owners, roleDeclarationOwner39247{logical, raw, set, file, function})
			}
		}
	}
	return owners, nil
}

func verifyRoleOwnerExtraction39247(t *testing.T, root, logical string) {
	t.Helper()
	result, err := ExtractWithResult(root, logical)
	if err != nil {
		t.Fatal("actual owner source extraction failed", logical, err)
	}
	if len(result.Generated) < 2 {
		t.Fatal("actual owner source partition did not emit helpers", logical)
	}
	for path, source := range result.Generated {
		if path != logical && physicalLines(source) > 75 {
			t.Fatal("emitted helper exceeded unchanged75-line cap", path)
		}
	}
}
