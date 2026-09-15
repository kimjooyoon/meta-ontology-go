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
				if ident, ok := selector.X.(*ast.Ident); ok && ident.Obj == nil {
					used[ident.Name] = true
				}
			}
			return true
		})
	}
	embedRequired := hasDirective(decls, "embed")
	linknameRequired := hasDirective(decls, "linkname")
	if !includeBlank && embedRequired && !hasBlankImport(list, "embed") {
		return nil, fail("validate-ast-imports", "select-imports", "EMBED_IMPORT_MISSING", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	if !includeBlank && linknameRequired && !hasBlankImport(list, "unsafe") {
		return nil, fail("validate-ast-imports", "select-imports", "LINKNAME_UNSAFE_IMPORT_MISSING", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	result := map[*ast.GenDecl][]*ast.ImportSpec{}
	for _, item := range list {
		if item.name == "." {
			return nil, fail("validate-ast-imports", "select-imports", "UNSUPPORTED_DOT_IMPORT", "KNOWN_CONTRADICTION", "report-contradiction", []string{item.path})
		}
		directiveImport := (embedRequired && item.path == "embed" && item.name == "_") ||
			(linknameRequired && item.path == "unsafe" && item.name == "_")
		if item.name == "_" && !includeBlank && !directiveImport {
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

func hasDirective(decls []ast.Decl, directive string) bool {
	prefix := "//go:" + directive
	for _, decl := range decls {
		var docs *ast.CommentGroup
		switch node := decl.(type) {
		case *ast.FuncDecl:
			docs = node.Doc
		case *ast.GenDecl:
			docs = node.Doc
		}
		if docs == nil {
			continue
		}
		for _, comment := range docs.List {
			text := strings.TrimSpace(comment.Text)
			if text == prefix || strings.HasPrefix(text, prefix+" ") || strings.HasPrefix(text, prefix+"\t") {
				return true
			}
		}
	}
	return false
}

func hasBlankImport(list []importSpec, path string) bool {
	for _, item := range list {
		if item.path == path && item.name == "_" {
			return true
		}
	}
	return false
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
