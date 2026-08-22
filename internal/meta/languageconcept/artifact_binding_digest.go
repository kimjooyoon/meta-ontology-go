package languageconcept

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io/fs"
	"sort"
)

func uniqueBindingPaths(concepts []Concept) []string {
	seen := map[string]bool{CatalogSourcePath: true}
	for _, concept := range concepts {
		for _, path := range concept.CodeBindings {
			seen[path] = true
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func observeBindingFile(repository fs.FS, path string, mode fs.FileMode, output hash.Hash, result *BindingObservation) {
	if !mode.IsRegular() {
		result.Unsupported++
		writeBindingRecord(output, "unsupported", path, []byte(mode.String()))
		return
	}
	content, err := fs.ReadFile(repository, path)
	if err != nil {
		result.Missing++
		writeBindingRecord(output, "unreadable", path, []byte(err.Error()))
		return
	}
	result.Files++
	result.Bytes += int64(len(content))
	writeBindingRecord(output, "file", path, content)
}

func writeBindingRecord(output hash.Hash, kind, path string, content []byte) {
	contentDigest := sha256.Sum256(content)
	_, _ = fmt.Fprintf(output, "%s\x00%s\x00%x\n", kind, path, contentDigest)
}
