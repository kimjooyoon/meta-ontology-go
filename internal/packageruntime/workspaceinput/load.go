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
	loaded, err := loadSources(reader, manifestPath, definition)
	if err != nil {
		// A prefix of the requested package graph may itself be executable.
		// Keep a failed read from masquerading as that complete workspace.
		return packageruntime.Manifest{}, err
	}
	return loaded, nil
}

func loadSources(reader Reader, manifestPath string, definition Definition) (packageruntime.Manifest, error) {
	result := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: definition.Entry}
	if err := validate(definition); err != nil {
		return result, err
	}
	if reader == nil {
		return result, fmt.Errorf("workspace source reader is required")
	}
	root := filepath.Dir(manifestPath)
	contents := make(map[string]string)
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
			content, err := readSource(reader, root, sourcePath, relative, contents)
			if err != nil {
				return result, err
			}
			size += len(content)
			if size > MaxSourceSetBytes {
				return result, fmt.Errorf("workspace source size exceeds its file or source-set limit")
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{Filename: filepath.ToSlash(relative), Content: content})
		}
		if err := packageruntime.CompleteSourceMetadata(&pkg, declared.Name, declared.Imports); err != nil {
			return result, err
		}
		result.Packages = append(result.Packages, pkg)
	}
	return result, nil
}

func readSource(reader Reader, root, authored, relative string, contents map[string]string) (string, error) {
	if content, known := contents[relative]; known {
		return content, nil
	}
	raw, err := reader.ReadFile(filepath.Join(root, relative))
	if err == nil && len(raw) > MaxSourceBytes {
		err = fmt.Errorf("input exceeds maximum size of %d bytes", MaxSourceBytes)
	}
	if err != nil {
		return "", fmt.Errorf("source %q: %w", authored, err)
	}
	content := string(raw)
	contents[relative] = content
	return content, nil
}
