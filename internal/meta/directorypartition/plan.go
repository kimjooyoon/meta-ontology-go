package directorypartition

import (
	"fmt"
	"path"
	"sort"
)

type Candidate struct {
	Subject                string `json:"subject"`
	IndicatorID            string `json:"indicator_id"`
	Limit                  int    `json:"limit"`
	EntryCount             int    `json:"entry_count"`
	BucketCount            int    `json:"bucket_count"`
	ProjectedDirectEntries int    `json:"projected_direct_entries"`
	Moves                  []Move `json:"moves"`
	Status                 string `json:"status"`
}

type Move struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Kind        string `json:"kind"`
	Bucket      int    `json:"bucket"`
}

func partitionTargets(source SourceMetrics) ([]SourceIndicator, int, int) {
	targets := make([]SourceIndicator, 0)
	applicable, rootExemptions := 0, 0
	for _, indicator := range source.Meta.Indicators {
		if indicator.Subject == "." && indicator.Applicability == "NOT_APPLICABLE" {
			rootExemptions++
		}
		if indicator.MetaOperation != "partition-directory" ||
			indicator.Applicability != "APPLICABLE" || !indicator.Blocking {
			continue
		}
		applicable++
		if !indicator.Satisfied {
			targets = append(targets, indicator)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Subject < targets[j].Subject })
	return targets, applicable, rootExemptions
}

func makeCandidate(source SourceMetrics, target SourceIndicator) (Candidate, error) {
	entries, limit := directEntries(source, target.Subject), target.Limit
	if limit <= 0 {
		limit = source.Meta.Policy.MaxDirectDirectoryEntries
	}
	if len(entries) != target.Value {
		return Candidate{}, fmt.Errorf("%s direct entry count is incoherent", target.Subject)
	}
	needed := (len(entries) + limit - 1) / limit
	if needed < 1 {
		needed = 1
	}
	buckets, status := needed, "REVIEW_REQUIRED"
	if buckets > limit {
		buckets, status = limit, "RECURSIVE_REVIEW_REQUIRED"
	}
	moves := make([]Move, 0, len(entries))
	for index, entry := range entries {
		bucket := index%buckets + 1
		base := fmt.Sprintf("_partition_%02d", bucket)
		moves = append(moves, Move{Source: entry.Path, Destination: path.Join(target.Subject, base, path.Base(entry.Path)), Kind: entry.Kind, Bucket: bucket})
	}
	return Candidate{Subject: target.Subject, IndicatorID: target.MetricID, Limit: limit, EntryCount: len(entries), BucketCount: buckets, ProjectedDirectEntries: buckets, Moves: moves, Status: status}, nil
}
