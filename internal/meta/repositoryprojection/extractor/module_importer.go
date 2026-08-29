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

type moduleImporter struct {
	root     string
	module   string
	fallback types.Importer
	packages map[string]*types.Package
	loading  map[string]bool
}

func newModuleImporter(root string) *moduleImporter {
	return &moduleImporter{root: root, module: readModulePath(root), fallback: importer.Default(),
		packages: map[string]*types.Package{}, loading: map[string]bool{}}
}

func (imports *moduleImporter) Import(path string) (*types.Package, error) {
	if package_, ok := imports.packages[path]; ok {
		return package_, nil
	}
	if !imports.local(path) {
		return imports.fallback.Import(path)
	}
	if imports.loading[path] {
		return nil, os.ErrInvalid
	}
	imports.loading[path] = true
	defer delete(imports.loading, path)
	files, err := imports.files(path)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(files))
	for _, name := range files {
		data, readErr := os.ReadFile(name)
		if readErr != nil {
			return nil, readErr
		}
		file, parseErr := parser.ParseFile(fset, name, data, parser.ParseComments)
		if parseErr != nil {
			return nil, parseErr
		}
		parsed = append(parsed, file)
	}
	checked, checkErr := (&types.Config{Importer: imports, Error: func(error) {}}).Check(path, fset, parsed, nil)
	if checkErr != nil {
		return nil, checkErr
	}
	if checked == nil {
		return nil, os.ErrInvalid
	}
	imports.packages[path] = checked
	return checked, nil
}

func (imports *moduleImporter) local(path string) bool {
	return imports.module != "" && (path == imports.module || strings.HasPrefix(path, imports.module+"/"))
}

func (imports *moduleImporter) files(path string) ([]string, error) {
	relative := strings.TrimPrefix(path, imports.module)
	directory := filepath.Join(imports.root, filepath.FromSlash(strings.TrimPrefix(relative, "/")))
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") ||
			!moduleBuildFile(directory, entry.Name()) {
			continue
		}
		result = append(result, filepath.Join(directory, entry.Name()))
	}
	return result, nil
}

func moduleBuildFile(directory, name string) bool {
	matched, err := build.Default.MatchFile(directory, name)
	return err == nil && matched
}

func readModulePath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}
