package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	return manifest{
		Schema: "gooo.repository-projection.v1", SourceSHA: sha,
		Proof: "axiomatic-foundation", Authority: "git-index-at-exact-head",
		Entries: entries,
	}
}

func decimalKey(hexID string) string {
	key := make([]byte, 0, len(hexID)*2)
	for _, digit := range []byte(hexID) {
		value := digit - '0'
		if digit >= 'a' {
			value = digit - 'a' + 10
		}
		key = append(key, '0'+value/10, '0'+value%10)
	}
	return string(key)
}

func writeEvidence(work string, report evidence) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	name := filepath.Join(work, "evidence.json")
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func requireBlockingZero(report evidence) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
