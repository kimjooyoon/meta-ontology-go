package directorypartition

import (
	"fmt"
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
		if directory.Path != "." && directory.Path != subject && path.Dir(directory.Path) == subject {
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
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Path != ordered[j].Path {
			return ordered[i].Path < ordered[j].Path
		}
		return ordered[i].Kind < ordered[j].Kind
	})
	return ordered
}

func makeCandidate(source SourceMetrics, target SourceIndicator) (Candidate, error) {
	entries, limit := directEntries(source, target.Subject), target.Limit
	if limit <= 0 {
		limit = source.Meta.Policy.MaxDirectDirectoryEntries
	}
	if len(entries) == 0 {
		return Candidate{}, fmt.Errorf("%s has no direct entries to partition", target.Subject)
	}
	needed := max((len(entries)+limit-1)/limit, 1)
	buckets, status := needed, "REVIEW_REQUIRED"
	if buckets > limit {
		buckets, status = limit, "RECURSIVE_REVIEW_REQUIRED"
	}
	moves := make([]Move, 0, len(entries))
	for index, entry := range entries {
		bucket := index%buckets + 1
		base := fmt.Sprintf("_partition_%02d", bucket)
		moves = append(moves, Move{
			Source: entry.Path, Destination: path.Join(target.Subject, base, path.Base(entry.Path)),
			Kind: entry.Kind, Bucket: bucket,
		})
	}
	return Candidate{
		Subject: target.Subject, IndicatorID: target.MetricID, Limit: limit,
		EntryCount: len(entries), BucketCount: buckets, ProjectedDirectEntries: buckets,
		Moves: moves, Status: status,
	}, nil
}
