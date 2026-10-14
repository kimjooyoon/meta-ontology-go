package extractor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// One type stage owns source snapshots and their Go compilation origins. The
// imported declarations have compiled bodies; the analyzed body remains fresh.
type moduleCompilationOrigin struct {
	ImportPath, BuildID, ExportSHA256, SourceSHA256, ContextSHA256 string
	SourceFiles                                                    int
}
type moduleCompiledPackage struct {
	Dir, ImportPath, BuildID, Export                                                                         string
	GoFiles, CgoFiles, CFiles, CXXFiles, MFiles, HFiles, FFiles, SFiles, SwigFiles, SwigCXXFiles, EmbedFiles []string
	Goroot, Incomplete                                                                                       bool
}
type moduleSourceInput struct {
	Name  string
	Bytes []byte
	Go    bool
}
type moduleCompiledInput struct {
	Package moduleCompiledPackage
	Inputs  []moduleSourceInput
	Digest  string
}

func (imports *moduleImporter) importCompiledDeclarations(path string) (*types.Package, error) {
	if !imports.local(path) && imports.standardLibrary(path) {
		pkg, err := imports.standard.Import(imports, path)
		if err != nil || pkg == nil {
			return nil, preserveModuleFailure(err)
		}
		imports.packages[path] = pkg
		return pkg, nil
	}
	if !imports.local(path) && imports.externalPackages >= modulePackageMax {
		return nil, moduleFailure(moduleCausePackageLimit)
	}
	if imports.loading[path] {
		return nil, os.ErrInvalid
	}
	if err := imports.ensureCompilationOrigin(path); err != nil {
		return nil, err
	}
	input, ok := imports.compilationInputs[path]
	if !ok {
		return nil, errModuleDependency
	}
	imports.loading[path] = true
	defer delete(imports.loading, path)
	set := token.NewFileSet()
	files := []*ast.File{}
	for _, source := range input.Inputs {
		if !source.Go {
			continue
		}
		file, err := parser.ParseFile(set, source.Name, source.Bytes, parser.ParseComments)
		if err != nil {
			return nil, errModuleDependency
		}
		files = append(files, file)
	}
	configuration := types.Config{Importer: imports, IgnoreFuncBodies: true, DisableUnusedImportCheck: true}
	pkg, err := configuration.Check(path, set, files, nil)
	if err != nil || pkg == nil {
		return nil, errModuleDependency
	}
	imports.packages[path] = pkg
	return pkg, nil
}

func sourceInputDigest(inputs []moduleSourceInput) string {
	type entry struct {
		Name, SHA256 string
		Go           bool
	}
	rows := make([]entry, 0, len(inputs))
	for _, input := range inputs {
		sum := sha256.Sum256(input.Bytes)
		rows = append(rows, entry{input.Name, hex.EncodeToString(sum[:]), input.Go})
	}
	raw, _ := json.Marshal(rows)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func exportFileDigest(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", errModuleDependency
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > moduleExportMax {
		return "", errModuleDependency
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, io.LimitReader(file, moduleExportMax+1)); err != nil {
		return "", errModuleDependency
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func compilationFileNames(pkg moduleCompiledPackage) []string {
	names := []string{}
	for _, group := range [][]string{pkg.GoFiles, pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles, pkg.FFiles, pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.EmbedFiles} {
		names = append(names, group...)
	}
	slices.Sort(names)
	return slices.Compact(names)
}
func normalizedPackageDirectory(directory string) (string, error) {
	if !filepath.IsAbs(directory) {
		return "", errModuleDependency
	}
	real, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", errModuleDependency
	}
	return real, nil
}
