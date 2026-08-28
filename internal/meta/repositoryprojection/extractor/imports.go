package extractor

import (
	"go/ast"
	"go/build"
	"os"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
)

func imports(file *ast.File) ([]importSpec, error) {
	var out []importSpec
	for _, node := range file.Decls {
		group, ok := node.(*ast.GenDecl)
		if !ok || group.Tok.String() != "import" {
			continue
		}
		for _, raw := range group.Specs {
			spec := raw.(*ast.ImportSpec)
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil || value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "..") {
				return nil, fail("validate-ast-imports", "parse-import", "UNSAFE_IMPORT_PATH", "KNOWN_CONTRADICTION", "report-contradiction", nil)
			}
			name := ""
			if spec.Name != nil {
				name = spec.Name.Name
			}
			out = append(out, importSpec{group, spec, value, name})
		}
	}
	return out, nil
}

func validateImports(root string, list []importSpec) error {
	module := moduleName(root)
	for _, item := range list {
		if item.name == "." {
			return fail("validate-ast-imports", "resolve-import", "UNSUPPORTED_DOT_IMPORT", "KNOWN_CONTRADICTION", "report-contradiction", []string{item.path})
		}
		if module != "" && (item.path == module || strings.HasPrefix(item.path, module+"/")) {
			local := root
			if item.path != module {
				local = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(item.path, module+"/")))
			}
			if !hasGoFile(local) {
				return fail("validate-ast-imports", "resolve-import", "UNRESOLVED_IMPORT", "DIRECT_MISSING", "restore-import-package", nil)
			}
			continue
		}
		if _, err := build.Default.Import(item.path, root, build.FindOnly); err != nil {
			return fail("validate-ast-imports", "resolve-import", "UNRESOLVED_IMPORT", "DIRECT_MISSING", "restore-import-package", nil)
		}
	}
	return nil
}

func moduleName(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}

func importName(item importSpec) string {
	if item.name != "" && item.name != "_" {
		return item.name
	}
	base := pathpkg.Base(item.path)
	if strings.HasPrefix(base, "v") && len(base) > 1 && base[1] >= '0' && base[1] <= '9' {
		base = pathpkg.Base(pathpkg.Dir(item.path))
	}
	return base
}
