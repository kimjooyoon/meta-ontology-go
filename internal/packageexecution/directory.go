package packageexecution

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const maxSourceBytes = 1 << 20

func LoadDirectory(directory string) ([]Source, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var sources []Source
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".gooo" {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return nil, fmt.Errorf("packageexecution: source %q is not a regular file", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		if len(data) > maxSourceBytes {
			return nil, fmt.Errorf("packageexecution: source %q exceeds %d bytes", entry.Name(), maxSourceBytes)
		}
		sources = append(sources, Source{Filename: entry.Name(), Content: string(data)})
	}
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	if len(sources) < 2 {
		return nil, fmt.Errorf("packageexecution: directory requires at least two .gooo files")
	}
	return sources, nil
}
