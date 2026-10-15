package extractor

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

func (imports *moduleImporter) captureCompilationSources(packages []moduleCompiledPackage) (map[string]moduleCompiledInput, error) {
	result := map[string]moduleCompiledInput{}
	sourceBytes := imports.compilationSourceBytes
	sourceFiles := imports.compilationSourceFiles
	external := imports.externalPackages
	for _, pkg := range packages {
		if pkg.Goroot {
			continue
		}
		if pkg.ImportPath == "" || pkg.Incomplete || len(pkg.CgoFiles) != 0 || len(pkg.GoFiles) == 0 {
			return nil, errModuleDependency
		}
		if _, exists := result[pkg.ImportPath]; exists {
			return nil, errModuleDependency
		}
		directory, err := normalizedPackageDirectory(pkg.Dir)
		if err != nil {
			return nil, err
		}
		if imports.local(pkg.ImportPath) {
			root, err := filepath.EvalSymlinks(imports.absoluteRoot())
			if err != nil {
				return nil, errModuleDependency
			}
			if _, inside := canonicalContainedPath(root, directory); !inside {
				return nil, errModuleDependency
			}
		} else {
			if _, known := imports.compilationInputs[pkg.ImportPath]; !known && external >= modulePackageMax {
				return nil, moduleFailure(moduleCausePackageLimit)
			}
			if _, known := imports.compilationInputs[pkg.ImportPath]; !known {
				external++
			}
		}
		pkg.Dir = directory
		input := moduleCompiledInput{Package: pkg}
		names := compilationFileNames(pkg)
		if _, retained := imports.compilationInputs[pkg.ImportPath]; !retained {
			if len(names) > moduleStageFilesMax-sourceFiles {
				return nil, moduleFailure(moduleCauseFileLimit)
			}
			sourceFiles += len(names)
		}
		for _, name := range names {
			if !fs.ValidPath(filepath.ToSlash(name)) || filepath.IsAbs(name) {
				return nil, errModuleDependency
			}
			full, contained := canonicalContainedPath(directory, filepath.Join(directory, filepath.FromSlash(name)))
			if !contained {
				return nil, errModuleDependency
			}
			data, err := readBoundedModuleSource(full)
			if err != nil {
				return nil, err
			}
			if _, retained := imports.compilationInputs[pkg.ImportPath]; !retained {
				if int64(len(data)) > moduleSourceMax-sourceBytes {
					return nil, moduleFailure(moduleCauseSourceLimit)
				}
				sourceBytes += int64(len(data))
			}
			input.Inputs = append(input.Inputs, moduleSourceInput{Name: full, Bytes: data, Go: slices.Contains(pkg.GoFiles, name)})
		}
		input.Digest = sourceInputDigest(input.Inputs)
		if previous, exists := imports.compilationInputs[pkg.ImportPath]; exists && previous.Digest != input.Digest {
			return nil, errModuleDependency
		}
		result[pkg.ImportPath] = input
	}
	return result, nil
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
