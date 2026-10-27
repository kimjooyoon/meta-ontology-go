package extractor

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// Make an implicit qualifier explicit only when its owning package declares a
// different name. Renderers can then use the actual binding at every split.
func bindDefaultImportAliases(root, logical string, file *ast.File) error {
	list, err := imports(file)
	if err != nil {
		return err
	}
	if err := validateImports(root, logical, list); err != nil {
		return err
	}
	owner, err := sourceModuleRoot(root, logical)
	if err != nil {
		return err
	}
	resolver := newModuleImporter(owner)
	module := moduleName(owner)
	for _, item := range list {
		if item.spec.Name != nil {
			continue
		}
		name, err := declaredImportName(owner, module, resolver, item.path)
		if err != nil {
			return fail("validate-ast-imports", "bind-package-name", "IMPORT_PACKAGE_NAME_UNAVAILABLE", "DIRECT_MISSING", "restore-import-package", []string{item.path})
		}
		if name != importName(item) {
			item.spec.Name = ast.NewIdent(name)
		}
	}
	return nil
}

func declaredImportName(owner, module string, resolver *moduleImporter, path string) (string, error) {
	directory := ""
	if module != "" && (path == module || strings.HasPrefix(path, module+"/")) {
		directory = filepath.Join(owner, filepath.FromSlash(strings.TrimPrefix(strings.TrimPrefix(path, module), "/")))
	} else if pkg, err := build.Default.Import(path, owner, build.FindOnly); err == nil && (pkg.Goroot || module == "") {
		directory = pkg.Dir
	}
	if directory != "" {
		pkg, err := build.Default.ImportDir(directory, 0)
		if err != nil {
			return "", err
		}
		return pkg.Name, nil
	}
	files, err := resolver.moduleFiles(path)
	if err != nil {
		return "", err
	}
	file, err := parser.ParseFile(token.NewFileSet(), files[0], nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return file.Name.Name, nil
}
