package main

import (
	"fmt"
	"go/format"
	"os"
)

func formatStaged(root string, buffers map[string][]byte) (map[string]stagedFile, error) {
	staged := make(map[string]stagedFile, len(buffers))
	for logical, data := range buffers {
		name, err := extractionPath(root, logical)
		if err != nil {
			return nil, err
		}
		formatted, err := format.Source(data)
		if err != nil {
			return nil, fmt.Errorf("format extraction %s: %w", logical, err)
		}
		if lines := extractionLines(formatted); lines > 75 {
			return nil, fmt.Errorf("extraction target %s remains at %d lines", logical, lines)
		}
		info, err := os.Stat(name)
		if err != nil {
			return nil, err
		}
		staged[logical] = stagedFile{name: name, data: formatted, mode: uint32(info.Mode().Perm())}
	}
	return staged, nil
}
