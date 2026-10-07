package main

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func loadPackageSources(reader SourceReader, manifestPath string, manifest workspaceManifest) (packageruntime.Manifest, error) {
	result := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: manifest.Entry}
	root := filepath.Dir(manifestPath)
	count, size := 0, 0
	for _, declared := range manifest.Packages {
		pkg := packageruntime.PackageSpec{Path: declared.Path, Name: declared.Name, Imports: append([]string(nil), declared.Imports...)}
		for _, sourcePath := range declared.Sources {
			relative, err := workspaceSourcePath(sourcePath)
			if err != nil {
				return result, err
			}
			count++
			if count > workspaceMaxSourceCount {
				return result, fmt.Errorf("workspace declares more than %d source files", workspaceMaxSourceCount)
			}
			content, err := readSource(reader, filepath.Join(root, relative))
			if err != nil {
				return result, fmt.Errorf("source %q: %w", sourcePath, err)
			}
			size += len(content)
			if len(content) > workspaceMaxSourceBytes || size > workspaceMaxSourceSetSize {
				return result, fmt.Errorf("workspace source size exceeds its file or source-set limit")
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{Filename: filepath.ToSlash(relative), Content: string(content)})
		}
		result.Packages = append(result.Packages, pkg)
	}
	return result, nil
}
