package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func writeBundle(output string, source, projection []byte) error {
	name := filepath.Base(output)
	if !strings.HasSuffix(name, ".generated.go") {
		return fmt.Errorf("projection output requires the .generated.go suffix")
	}
	parts, err := generator.PartitionProjection(projection, 75)
	if err != nil {
		return err
	}
	stem, directory := strings.TrimSuffix(name, ".generated.go"), filepath.Dir(output)
	manifest := projectionbundle.Manifest{Schema: projectionbundle.Schema, SourceDigest: projectionbundle.Digest(source)}
	files := make(map[string][]byte, len(parts))
	for i, part := range parts {
		path := name
		if i > 0 {
			path = fmt.Sprintf("%s_part%02d.generated.go", stem, i+1)
		}
		files[path] = part.Source
		manifest.Members = append(manifest.Members, projectionbundle.Member{
			Path: path, Digest: projectionbundle.Digest(part.Source), RegionIDs: part.RegionIDs,
		})
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if _, err := projectionbundle.Decode(raw); err != nil {
		return err
	}
	manifestPath := filepath.Join(directory, stem+".projection.json")
	obsolete, err := obsoleteMembers(manifestPath, manifest)
	if err != nil {
		return err
	}
	for _, member := range manifest.Members {
		if err := writeProjection(filepath.Join(directory, member.Path), files[member.Path]); err != nil {
			return err
		}
	}
	for _, path := range obsolete {
		if err := os.Remove(filepath.Join(directory, path)); err != nil {
			return err
		}
	}
	// The manifest commits the file set last. Interrupted updates remain unbound.
	return writeProjection(manifestPath, append(raw, '\n'))
}

func obsoleteMembers(path string, next projectionbundle.Manifest) ([]string, error) {
	return projectionbundle.ObsoleteMembers(path, next)
}
