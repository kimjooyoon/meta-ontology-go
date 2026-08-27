package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
)

func ObservePolicy(source, generated string) PolicyEvidence {
	evidence := PolicyEvidence{}
	raw, err := os.ReadFile(source)
	if err == nil {
		evidence.SourceAvailable = true
		evidence.SourceDigest = digestBytes(raw)
	}
	digest, count, err := digestTree(generated)
	if err == nil && count > 0 {
		evidence.GeneratedAvailable = true
		evidence.GeneratedDigest = digest
	}
	return evidence
}

func digestTree(root string) (string, int, error) {
	hash := sha256.New()
	count := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash.Write([]byte(filepath.ToSlash(relative)))
		hash.Write([]byte{0})
		hash.Write(raw)
		hash.Write([]byte{0})
		count++
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), count, nil
}
