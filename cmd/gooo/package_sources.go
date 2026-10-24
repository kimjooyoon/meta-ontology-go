package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func loadPackageSources(reader SourceReader, manifestPath string, manifest workspaceManifest) (packageruntime.Manifest, error) {
	return workspaceinput.Load(reader, manifestPath, manifest)
}
