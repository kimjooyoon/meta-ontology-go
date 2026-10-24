package extractionplan

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
)

func formatStaged(root string, buffers map[string][]byte, created map[string]bool) (map[string]File, error) {
	staged := make(map[string]File, len(buffers))
	for _, logical := range slices.Sorted(maps.Keys(buffers)) {
		data := buffers[logical]
		name, err := extractionPath(root, logical)
		if err != nil {
			return nil, err
		}
		formatted, err := format.Source(data)
		if err != nil {
			return nil, fmt.Errorf("format extraction %s: %w", logical, err)
		}
		if extractionLines(formatted) > 75 {
			return nil, fmt.Errorf("extraction target %s remains at %d lines", logical, extractionLines(formatted))
		}
		mode := uint32(0o644)
		if !created[logical] {
			info, err := os.Stat(name)
			if err != nil {
				return nil, err
			}
			mode = uint32(info.Mode().Perm())
		}
		staged[logical] = File{Name: name, Data: formatted, Mode: mode, Created: created[logical]}
	}
	return staged, nil
}

func Lines(data []byte) int { return extractionLines(data) }
