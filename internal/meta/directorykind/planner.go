package directorykind

import (
	"fmt"
	"path"
	"sort"
)

type directEntry struct {
	Path     string
	Kind     string
	Language string
}

func directEntries(source SourceMetrics, subject string) []directEntry {
	entries := make(map[string]directEntry)
	for _, directory := range source.Directories {
		if directory.Path != "." && directory.Path != subject && path.Dir(directory.Path) == subject {
			entries[directory.Path] = directEntry{Path: directory.Path, Kind: "directory"}
		}
	}
	for _, file := range source.Files {
		if path.Dir(file.Path) == subject {
			entries[file.Path] = directEntry{Path: file.Path, Kind: "file", Language: file.Language}
		}
	}
	ordered := make([]directEntry, 0, len(entries))
	for _, entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	return ordered
}

func availableGroup(subject, base string, occupied map[string]bool) string {
	for index := 1; ; index++ {
		name := base
		if index > 1 {
			name = fmt.Sprintf("%s_%02d", base, index)
		}
		candidate := path.Join(subject, name)
		if !occupied[candidate] {
			occupied[candidate] = true
			return candidate
		}
	}
}

func makeCandidate(source SourceMetrics, target SourceIndicator) (Candidate, error) {
	if target.Subject == "." {
		return Candidate{}, fmt.Errorf("project root kind topology is exempt")
	}
	metric, ok := directoryMetric(source, target.Subject)
	if !ok || metric.DirectFolders == 0 || metric.DirectFiles == 0 {
		return Candidate{}, fmt.Errorf("%s is not a mixed directory", target.Subject)
	}
	entries := directEntries(source, target.Subject)
	if len(entries) != metric.DirectFolders+metric.DirectFiles {
		return Candidate{}, fmt.Errorf("%s direct entry evidence is incomplete", target.Subject)
	}
	occupied := make(map[string]bool, len(entries)+2)
	for _, entry := range entries {
		occupied[entry.Path] = true
	}
	directories := availableGroup(target.Subject, "_kind_directories", occupied)
	files := availableGroup(target.Subject, "_kind_files", occupied)
	moves := make([]Move, 0, len(entries))
	folderCount, fileCount := 0, 0
	for _, entry := range entries {
		destination := files
		if entry.Kind == "directory" {
			destination, folderCount = directories, folderCount+1
		} else {
			fileCount++
		}
		moves = append(moves, Move{Source: entry.Path, Destination: path.Join(destination, path.Base(entry.Path)),
			EntryKind: entry.Kind, Language: entry.Language})
	}
	groups := []Group{{Kind: "directory", Destination: directories, EntryCount: folderCount},
		{Kind: "file", Destination: files, EntryCount: fileCount}}
	return Candidate{Subject: target.Subject, IndicatorID: target.MetricID, ObservedKinds: target.Value,
		EntryCount: len(entries), GroupCount: len(groups), Groups: groups, Moves: moves, Status: "REVIEW_REQUIRED"}, nil
}
