package extractor

import (
	"go/ast"
	"os"
	"slices"
	"sort"
	"strings"
)

func selectedImports(decls []ast.Decl, list []importSpec, includeBlank bool) (map[*ast.GenDecl][]*ast.ImportSpec, error) {
	used := map[string]bool{}
	for _, decl := range decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					used[ident.Name] = true
				}
			}
			return true
		})
	}
	result := map[*ast.GenDecl][]*ast.ImportSpec{}
	for _, item := range list {
		if item.name == "." {
			return nil, fail("validate-ast-imports", "select-imports", "UNSUPPORTED_DOT_IMPORT", "KNOWN_CONTRADICTION", "report-contradiction", []string{item.path})
		}
		if item.name == "_" && !includeBlank {
			continue
		}
		if item.name == "_" || used[importName(item)] {
			result[item.group] = append(result[item.group], item.spec)
		}
	}
	for _, specs := range result {
		sort.SliceStable(specs, func(i, j int) bool { return specs[i].Pos() < specs[j].Pos() })
	}
	return result, nil
}

func hasGoFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}
