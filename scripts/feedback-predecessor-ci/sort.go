package main

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func sortCandidates(candidates []feedbackpredecessor.Candidate) {
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		return left.RunID < right.RunID ||
			left.RunID == right.RunID && left.ArtifactID < right.ArtifactID
	})
}
