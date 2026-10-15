package extractor

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
)

func (imports *moduleImporter) listCompilationClosure(path string, export bool) ([]moduleCompiledPackage, error) {
	args := append([]string{"list"}, imports.moduleListCommonArgs()...)
	args = append(args, "-deps", "-json=Dir,ImportPath,BuildID,Export,GoFiles,CgoFiles,CFiles,CXXFiles,MFiles,HFiles,FFiles,SFiles,SwigFiles,SwigCXXFiles,EmbedFiles,Goroot,Incomplete")
	if export {
		args = append(args, "-export")
	}
	args = append(args, path)
	output, err := imports.runGoList(args)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	packages := []moduleCompiledPackage{}
	ids := map[string]bool{}
	found := false
	for {
		var pkg moduleCompiledPackage
		err = decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil || pkg.ImportPath == "" || pkg.Incomplete || ids[pkg.ImportPath] {
			return nil, errModuleDependency
		}
		if len(packages) >= moduleListPackagesMax {
			return nil, moduleFailure(moduleCausePackageLimit)
		}
		ids[pkg.ImportPath] = true
		found = found || pkg.ImportPath == path
		packages = append(packages, pkg)
	}
	if !found {
		return nil, errModuleDependency
	}
	return packages, nil
}

func (imports *moduleImporter) ensureCompilationOrigin(path string) error {
	if _, exists := imports.compilationInputs[path]; exists {
		return nil
	}
	beforeScope, err := standardImportScope(imports.root)
	if err != nil {
		return errModuleDependency
	}
	discovered, err := imports.listCompilationClosure(path, false)
	if err != nil {
		return err
	}
	before, err := imports.captureCompilationSources(discovered)
	if err != nil {
		return err
	}
	compiled, err := imports.listCompilationClosure(path, true)
	if err != nil {
		return err
	}
	after, err := imports.captureCompilationSources(compiled)
	if err != nil {
		return err
	}
	afterScope, err := standardImportScope(imports.root)
	if err != nil || beforeScope != afterScope || !sameCompilationInputs(before, after) {
		return errModuleDependency
	}
	if !imports.currentCompilationInputs(after) {
		return errModuleDependency
	}
	origins, err := compilationOrigins(after, afterScope)
	if err != nil {
		return err
	}
	if err = imports.retainStandardExportOrigins(compiled); err != nil {
		return err
	}
	imports.retainCompilationOrigins(after, origins)
	return nil
}

func (imports *moduleImporter) retainStandardExportOrigins(packages []moduleCompiledPackage) error {
	loaded := map[string]string{}
	for _, pkg := range packages {
		if !pkg.Goroot || pkg.ImportPath == "unsafe" {
			continue
		}
		if _, err := exportFileDigest(pkg.Export); err != nil {
			return err
		}
		loaded[pkg.ImportPath] = pkg.Export
	}
	added := 0
	for path := range loaded {
		if _, exists := imports.exportPaths[path]; !exists {
			added++
		}
	}
	if added > moduleListPackagesMax-imports.exportPackages {
		return moduleFailure(moduleCausePackageLimit)
	}
	maps.Copy(imports.exportPaths, loaded)
	imports.exportPackages += added
	return nil
}
