package extractor

import (
	"go/ast"
	"strconv"
)

// The projected package owns one import population. Observe that population's
// compilation closure together, while keeping target AST/type checking fresh.
// Standard-only stages retain the existing shared standard-import reader.
func (imports *moduleImporter) prepareCompilationImports(files []*ast.File) error {
	roots := []string{}
	project := false
	for _, file := range files {
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil || !validModuleImportPath(path) {
				return errModuleDependency
			}
			if path == "unsafe" {
				continue
			}
			roots = append(roots, path)
			if imports.local(path) || !imports.standardLibrary(path) {
				project = true
			}
		}
	}
	if !project {
		return nil
	}
	return imports.ensureCompilationRoots(roots)
}
