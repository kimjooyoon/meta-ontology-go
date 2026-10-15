package extractor

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

type compilationCapture struct {
	imports         *moduleImporter
	inputs          map[string]moduleCompiledInput
	bytes           int64
	files, external int
}

func (imports *moduleImporter) captureCompilationSources(packages []moduleCompiledPackage) (map[string]moduleCompiledInput, error) {
	capture := compilationCapture{imports: imports, inputs: map[string]moduleCompiledInput{},
		bytes: imports.compilationSourceBytes, files: imports.compilationSourceFiles, external: imports.externalPackages}
	for _, pkg := range packages {
		if pkg.Goroot {
			continue
		}
		if err := capture.packageInput(pkg); err != nil {
			return nil, err
		}
	}
	return capture.inputs, nil
}

func (capture *compilationCapture) packageInput(pkg moduleCompiledPackage) error {
	if pkg.ImportPath == "" || pkg.Incomplete || len(pkg.CgoFiles) != 0 || len(pkg.GoFiles) == 0 {
		return errModuleDependency
	}
	if _, exists := capture.inputs[pkg.ImportPath]; exists {
		return errModuleDependency
	}
	directory, err := capture.directory(pkg)
	if err != nil {
		return err
	}
	pkg.Dir = directory
	input := moduleCompiledInput{Package: pkg}
	names := compilationFileNames(pkg)
	if err := capture.reserveFiles(pkg.ImportPath, len(names)); err != nil {
		return err
	}
	for _, name := range names {
		source, err := capture.sourceInput(pkg, name)
		if err != nil {
			return err
		}
		input.Inputs = append(input.Inputs, source)
	}
	input.Digest = sourceInputDigest(input.Inputs)
	if previous, exists := capture.imports.compilationInputs[pkg.ImportPath]; exists && previous.Digest != input.Digest {
		return errModuleDependency
	}
	capture.inputs[pkg.ImportPath] = input
	return nil
}

func (capture *compilationCapture) directory(pkg moduleCompiledPackage) (string, error) {
	directory, err := normalizedPackageDirectory(pkg.Dir)
	if err != nil {
		return "", err
	}
	if capture.imports.local(pkg.ImportPath) {
		root, err := filepath.EvalSymlinks(capture.imports.absoluteRoot())
		if err != nil {
			return "", errModuleDependency
		}
		if _, inside := canonicalContainedPath(root, directory); !inside {
			return "", errModuleDependency
		}
	} else if _, retained := capture.imports.compilationInputs[pkg.ImportPath]; !retained {
		if capture.external >= modulePackageMax {
			return "", moduleFailure(moduleCausePackageLimit)
		}
		capture.external++
	}
	return directory, nil
}

func (capture *compilationCapture) reserveFiles(path string, count int) error {
	if _, retained := capture.imports.compilationInputs[path]; !retained {
		if count > moduleStageFilesMax-capture.files {
			return moduleFailure(moduleCauseFileLimit)
		}
		capture.files += count
	}
	return nil
}

func (capture *compilationCapture) sourceInput(pkg moduleCompiledPackage, name string) (moduleSourceInput, error) {
	if !fs.ValidPath(filepath.ToSlash(name)) || filepath.IsAbs(name) {
		return moduleSourceInput{}, errModuleDependency
	}
	full, inside := canonicalContainedPath(pkg.Dir, filepath.Join(pkg.Dir, filepath.FromSlash(name)))
	if !inside {
		return moduleSourceInput{}, errModuleDependency
	}
	data, err := readBoundedModuleSource(full)
	if err != nil {
		return moduleSourceInput{}, err
	}
	if _, retained := capture.imports.compilationInputs[pkg.ImportPath]; !retained {
		if int64(len(data)) > moduleSourceMax-capture.bytes {
			return moduleSourceInput{}, moduleFailure(moduleCauseSourceLimit)
		}
		capture.bytes += int64(len(data))
	}
	return moduleSourceInput{Name: full, Bytes: data, Go: slices.Contains(pkg.GoFiles, name)}, nil
}

func sameCompilationInputs(left, right map[string]moduleCompiledInput) bool {
	if len(left) != len(right) {
		return false
	}
	for path, before := range left {
		after, ok := right[path]
		if !ok || before.Digest != after.Digest || before.Package.Dir != after.Package.Dir {
			return false
		}
		if !slices.Equal(before.Package.GoFiles, after.Package.GoFiles) || !slices.Equal(before.Package.CgoFiles, after.Package.CgoFiles) {
			return false
		}
	}
	return true
}

func compilationOrigins(inputs map[string]moduleCompiledInput, scope string) (map[string]moduleCompilationOrigin, error) {
	origins := map[string]moduleCompilationOrigin{}
	for path, input := range inputs {
		if input.Package.BuildID == "" || input.Package.Export == "" {
			return nil, errModuleDependency
		}
		exportSHA, err := exportFileDigest(input.Package.Export)
		if err != nil {
			return nil, err
		}
		origins[path] = moduleCompilationOrigin{ImportPath: path, BuildID: input.Package.BuildID, ExportSHA256: exportSHA,
			SourceSHA256: input.Digest, ContextSHA256: scope, SourceFiles: len(input.Inputs)}
	}
	return origins, nil
}

func (imports *moduleImporter) retainCompilationOrigins(inputs map[string]moduleCompiledInput, origins map[string]moduleCompilationOrigin) {
	if imports.compilationInputs == nil {
		imports.compilationInputs = map[string]moduleCompiledInput{}
		imports.compilationOrigins = map[string]moduleCompilationOrigin{}
	}
	for path, input := range inputs {
		if _, exists := imports.compilationInputs[path]; exists {
			continue
		}
		for _, source := range input.Inputs {
			imports.compilationSourceBytes += int64(len(source.Bytes))
		}
		imports.compilationSourceFiles += len(input.Inputs)
		imports.compilationInputs[path] = input
		imports.compilationOrigins[path] = origins[path]
		if !imports.local(path) {
			imports.externalPackages++
		}
	}
}

// Revalidation reads source bytes; it never makes a missing compilation origin
// eligible for a declaration-only type read.
func (imports *moduleImporter) currentCompilationInputs(inputs map[string]moduleCompiledInput) bool {
	for _, input := range inputs {
		for _, source := range input.Inputs {
			data, err := os.ReadFile(source.Name)
			if err != nil || !slices.Equal(data, source.Bytes) {
				return false
			}
		}
	}
	return true
}
