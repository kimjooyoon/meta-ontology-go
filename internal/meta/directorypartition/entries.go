package directorypartition

import (
	"path"
	"sort"
)

type directEntry struct {
	Path string
	Kind string
}

func directEntries(source SourceMetrics, subject string) []directEntry {
	entries := make(map[string]directEntry)
	for _, directory := range source.Directories {
		if directory.Path == "." || directory.Path == subject {
			continue
		}
		if path.Dir(directory.Path) == subject {
			entries[directory.Path] = directEntry{Path: directory.Path, Kind: "directory"}
		}
	}
	for _, file := range source.Files {
		if path.Dir(file.Path) == subject {
			entries[file.Path] = directEntry{Path: file.Path, Kind: file.Language}
		}
	}
	ordered := make([]directEntry, 0, len(entries))
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].Path != ordered[right].Path {
			return ordered[left].Path < ordered[right].Path
		}
		return ordered[left].Kind < ordered[right].Kind
	})
	return ordered
}
