package projectionbundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// ObsoleteMembers returns only files still owned by the previous manifest.
// A changed or foreign member remains an explicit conflict.
func ObsoleteMembers(path string, next Manifest) ([]string, error) {
	if len(next.Members) == 0 {
		return nil, fmt.Errorf("next projection manifest has no members")
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	previous, err := Decode(raw)
	if err != nil {
		return nil, err
	}
	if previous.Members[0].Path != next.Members[0].Path {
		return nil, fmt.Errorf("previous projection manifest belongs to another output")
	}
	keep := map[string]bool{}
	for _, member := range next.Members {
		keep[member.Path] = true
	}
	var obsolete []string
	for _, member := range previous.Members {
		if keep[member.Path] {
			continue
		}
		content, err := os.ReadFile(filepath.Join(filepath.Dir(path), member.Path))
		if err != nil || Digest(content) != member.Digest {
			return nil, fmt.Errorf("obsolete projection member %s is not owned by the previous manifest", member.Path)
		}
		obsolete = append(obsolete, member.Path)
	}
	return obsolete, nil
}

// WriteAtomic replaces one generated member. The caller publishes its manifest
// last; this function alone does not commit a multi-file transaction.
func WriteAtomic(path string, raw []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".gooo-projection-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0644); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
