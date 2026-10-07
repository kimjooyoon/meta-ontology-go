package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func readPackageAssemblyPolicy(reader SourceReader, path string) (*packageruntime.Manifest, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := readSource(reader, path)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxInputBytes {
		return nil, inputLimitError(maxInputBytes)
	}
	manifest, err := decodeWorkspaceManifest(raw)
	if err != nil {
		return nil, err
	}
	runtimeManifest, err := loadPackageSources(reader, path, manifest)
	return &runtimeManifest, err
}
