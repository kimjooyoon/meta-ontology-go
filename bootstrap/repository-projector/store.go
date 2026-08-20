package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

func buildManifest(sha string, files []trackedFile,
	objects map[string]*storedObject) manifest {
	entries := make([]manifestEntry, 0, len(files))
	for _, file := range files {
		entries = append(entries, manifestEntry{
			Logical: file.logical, Backing: objects[file.objectSHA].backing,
			ObjectSHA: file.objectSHA, ContentSHA: contentHash(file.data),
			Kind: file.kind, Language: file.language,
			Mode: file.mode, Lines: file.lines,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Logical < entries[j].Logical })
	return manifest{
		Schema: "gooo.repository-projection.v1", SourceSHA: sha,
		Proof: "axiomatic-foundation", Authority: "git-index-at-exact-head",
		Entries: entries,
	}
}

func writeStore(work string, model manifest,
	objects map[string]*storedObject) (string, error) {
	root := filepath.Join(work, "stored")
	names := make([]string, 0, len(objects))
	for name := range objects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		object := objects[name]
		target := filepath.Join(root, filepath.FromSlash(object.backing))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, object.data, 0o644); err != nil {
			return "", err
		}
	}
	catalog := filepath.Join(root, "catalog")
	if err := os.MkdirAll(catalog, 0o755); err != nil {
		return "", err
	}
	encoded, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(filepath.Join(catalog, "manifest.json"), encoded, 0o644); err != nil {
		return "", err
	}
	return root, nil
}
