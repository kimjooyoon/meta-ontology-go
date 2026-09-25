package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	projectionevidence "github.com/kimjooyoon/meta-ontology-go/bootstrap/repository-projector/evidence"
)

func buildManifest(sha string, files []trackedFile,
	objects map[string]*storedObject) manifest {
	entries := make([]manifestEntry, 0, len(files))
	for _, file := range files {
		backing := file.backing
		if backing == "" {
			backing = objects[file.objectSHA].backing
		}
		entries = append(entries, manifestEntry{
			Logical: file.logical, Backing: backing,
			ObjectSHA: file.objectSHA, ContentSHA: contentHash(file.data),
			Kind: file.kind, Language: file.language,
			Mode: file.mode, Lines: file.lines,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Logical < entries[j].Logical })
	return manifest{Schema: "gooo.repository-projection.v1", SourceSHA: sha, Proof: "axiomatic-foundation", Authority: "git-index-at-exact-head", Entries: entries}
}

func writeEvidence(work string, report projectionevidence.Report) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	name := filepath.Join(work, "evidence.json")
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func restorePhysical(settings config) error {
	identity, physical, work, err := physicalPaths(settings)
	if err != nil {
		return err
	}
	if err := verifyGitIdentity(identity, settings.expectedSHA); err != nil {
		return err
	}
	if err := prepareWork(identity, work); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(physical, "projection", "catalog", "manifest.json"))
	if err != nil {
		return err
	}
	var model manifest
	if err := json.Unmarshal(data, &model); err != nil {
		return err
	}
	if model.Schema != "gooo.repository-projection.v1" || model.SourceSHA == "" {
		return fmt.Errorf("stored projection manifest is not authoritative")
	}
	loss, err := materialize(physical, filepath.Join(work, "materialized"), model)
	if err != nil {
		return err
	}
	if loss != 0 {
		return fmt.Errorf("stored projection roundtrip loss=%d", loss)
	}
	fmt.Printf("repository-projector: restored=%d source=%s\n", len(model.Entries), model.SourceSHA)
	return nil
}
