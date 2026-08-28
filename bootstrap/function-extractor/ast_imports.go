package main

import (
	"go/ast"
	"go/build"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
)

func genericImports(file *ast.File) ([]astImport, error) {
	var out []astImport
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok.String() != "import" {
			continue
		}
		for _, raw := range group.Specs {
			spec := raw.(*ast.ImportSpec)
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil || value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "..") {
				return nil, extractionError("validate-ast-imports", "parse-import", "UNSAFE_IMPORT_PATH", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
			}
			name := ""
			if spec.Name != nil {
				name = spec.Name.Name
			}
			out = append(out, astImport{decl: group, spec: spec, path: value, name: name})
		}
	}
	return out, nil
}

func validateGenericImports(root string, imports []astImport) error {
	module := genericModule(root)
	for _, item := range imports {
		if item.name == "." {
			return extractionError("validate-ast-imports", "resolve-import", "UNSUPPORTED_DOT_IMPORT", "KNOWN_CONTRADICTION", "report-contradiction", []string{item.path})
		}
		if module != "" && (item.path == module || strings.HasPrefix(item.path, module+"/")) {
			local := root
			if item.path != module {
				local = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(item.path, module+"/")))
			}
			if !genericHasGoFile(local) {
				return extractionError("validate-ast-imports", "resolve-import", "UNRESOLVED_IMPORT", "DIRECT_MISSING", "restore-import-package", []string{})
			}
			continue
		}
		if _, err := build.Default.Import(item.path, root, build.FindOnly); err != nil {
			return extractionError("validate-ast-imports", "resolve-import", "UNRESOLVED_IMPORT", "DIRECT_MISSING", "restore-import-package", []string{})
		}
	}
	return nil
}

func genericImportName(item astImport) string {
	if item.name != "" && item.name != "_" {
		return item.name
	}
	base := pathpkg.Base(item.path)
	if strings.HasPrefix(base, "v") && len(base) > 1 {
		if base[1] >= '0' && base[1] <= '9' {
			base = pathpkg.Base(pathpkg.Dir(item.path))
		}
	}
	return base
}
