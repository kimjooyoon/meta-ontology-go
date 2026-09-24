package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementCausalFrontierObservationSchema = "gooo.self-improvement.causal-frontier.v1"

const (
	SelfImprovementCausalFrontierSelected = "SELECTED"
	SelfImprovementCausalFrontierUnknown  = "UNKNOWN"
)

type SelfImprovementCausalFrontierCandidate struct {
	ID              string `json:"id"`
	State           string `json:"state"`
	EvidenceCount   int    `json:"evidenceCount"`
	DependencyCount int    `json:"dependencyCount"`
	Digest          string `json:"digest"`
}

type SelfImprovementCausalFrontierObservation struct {
	Schema          string `json:"schema"`
	CandidateCount  int    `json:"candidateCount"`
	SelectedID      string `json:"selectedId"`
	SelectedState   string `json:"selectedState"`
	SelectedDigest  string `json:"selectedDigest"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	NonAuthorizing  bool   `json:"nonAuthorizing"`
	Digest          string `json:"digest"`
}

func ObserveSelfImprovementCausalFrontier(
	candidates []SelfImprovementCausalFrontierCandidate,
) SelfImprovementCausalFrontierObservation {
	observation := SelfImprovementCausalFrontierObservation{
		Schema:         SelfImprovementCausalFrontierObservationSchema,
		CandidateCount: len(candidates),
		NonAuthorizing: true,
	}

	var selected *SelfImprovementCausalFrontierCandidate
	for _, candidate := range candidates {
		if !validSelfImprovementCausalFrontierCandidate(candidate) {
			continue
		}
		current := candidate
		if selected == nil || beforeSelfImprovementCausalFrontier(current, *selected) {
			selected = &current
		}
	}

	if selected == nil {
		observation.Status = SelfImprovementCausalFrontierUnknown
		observation.Reason = "CAUSAL_FRONTIER_EMPTY"
		observation.Digest = digestSelfImprovementCausalFrontier(observation)
		return observation
	}

	observation.SelectedID = selected.ID
	observation.SelectedState = selected.State
	observation.SelectedDigest = selected.Digest
	observation.Status = SelfImprovementCausalFrontierSelected
	observation.Reason = "CAUSAL_FRONTIER_SELECTED"
	observation.Digest = digestSelfImprovementCausalFrontier(observation)
	return observation
}

func validSelfImprovementCausalFrontierCandidate(
	candidate SelfImprovementCausalFrontierCandidate,
) bool {
	switch candidate.State {
	case "REFUTED", "UNKNOWN", "CLOSED":
	default:
		return false
	}
	return candidate.ID != "" &&
		candidate.EvidenceCount >= 0 &&
		candidate.DependencyCount >= 0 &&
		candidate.Digest != ""
}

func beforeSelfImprovementCausalFrontier(
	left SelfImprovementCausalFrontierCandidate,
	right SelfImprovementCausalFrontierCandidate,
) bool {
	leftRank := selfImprovementCausalFrontierRank(left.State)
	rightRank := selfImprovementCausalFrontierRank(right.State)
	if leftRank != rightRank {
		return leftRank > rightRank
	}
	if left.DependencyCount != right.DependencyCount {
		return left.DependencyCount < right.DependencyCount
	}
	if left.EvidenceCount != right.EvidenceCount {
		return left.EvidenceCount < right.EvidenceCount
	}
	if left.ID != right.ID {
		return left.ID < right.ID
	}
	return left.Digest < right.Digest
}

func selfImprovementCausalFrontierRank(state string) int {
	switch state {
	case "REFUTED":
		return 3
	case "UNKNOWN":
		return 2
	case "CLOSED":
		return 1
	default:
		return 0
	}
}

func digestSelfImprovementCausalFrontier(
	observation SelfImprovementCausalFrontierObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}