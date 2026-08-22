package languageconcept

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io/fs"
)

func observeBindings(repository fs.FS, concepts []Concept) BindingObservation {
	paths := uniqueBindingPaths(concepts)
	output := sha256.New()
	result := BindingObservation{Paths: len(paths)}
	for _, path := range paths {
		observeBindingPath(repository, path, output, &result)
	}
	result.Digest = "sha256:" + hex.EncodeToString(output.Sum(nil))
	return result
}

func observeBindingPath(repository fs.FS, root string, output hash.Hash, result *BindingObservation) {
	info, err := fs.Stat(repository, root)
	if err != nil {
		result.Missing++
		writeBindingRecord(output, "missing", root, []byte(err.Error()))
		return
	}
	if !info.IsDir() {
		observeBindingFile(repository, root, info.Mode(), output, result)
		return
	}
	writeBindingRecord(output, "directory", root, nil)
	_ = fs.WalkDir(repository, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Missing++
			writeBindingRecord(output, "unreadable", path, []byte(walkErr.Error()))
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		entryInfo, infoErr := entry.Info()
		if infoErr != nil {
			result.Missing++
			writeBindingRecord(output, "unreadable", path, []byte(infoErr.Error()))
			return nil
		}
		observeBindingFile(repository, path, entryInfo.Mode(), output, result)
		return nil
	})
}
