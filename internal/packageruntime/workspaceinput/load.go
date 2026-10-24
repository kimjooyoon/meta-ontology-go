package workspaceinput

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

// Reader supplies current file bytes. Filesystem ownership and symlink handling
// stay with the host; Load bounds returned bytes and resolves source metadata.
type Reader interface{ ReadFile(string) ([]byte, error) }

func Load(reader Reader, manifestPath string, definition Definition) (packageruntime.Manifest, error) {
	result := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: definition.Entry}
	if err := validate(definition); err != nil {
		return result, err
	}
	if reader == nil {
		return result, fmt.Errorf("workspace source reader is required")
	}
	root := filepath.Dir(manifestPath)
	count, size := 0, 0
	for _, declared := range definition.Packages {
		pkg := packageruntime.PackageSpec{Path: declared.Path, Name: declared.Name, Imports: append([]string(nil), declared.Imports...)}
		for _, sourcePath := range declared.Sources {
			relative, err := SourcePath(sourcePath)
			if err != nil {
				return result, err
			}
			count++
			if count > MaxSourceCount {
				return result, fmt.Errorf("workspace declares more than %d source files", MaxSourceCount)
			}
			content, err := reader.ReadFile(filepath.Join(root, relative))
			if err == nil && len(content) > MaxSourceBytes {
				err = fmt.Errorf("input exceeds maximum size of %d bytes", MaxSourceBytes)
			}
			if err != nil {
				return result, fmt.Errorf("source %q: %w", sourcePath, err)
			}
			size += len(content)
			if size > MaxSourceSetBytes {
				return result, fmt.Errorf("workspace source size exceeds its file or source-set limit")
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{Filename: filepath.ToSlash(relative), Content: string(content)})
		}
		if err := packageruntime.CompleteSourceMetadata(&pkg, declared.Name, declared.Imports); err != nil {
			return result, err
		}
		result.Packages = append(result.Packages, pkg)
	}
	return result, nil
}
