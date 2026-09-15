package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"slices"
	"strings"
)

func classifyInferencePathCode(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "stable-id-collision"), strings.Contains(message, "duplicate"):
		return CodeDuplicate
	case strings.Contains(message, "stale-evidence"):
		return CodeDigestMismatch
	case strings.Contains(message, "snapshot"):
		return CodeStaleSnapshot
	case strings.Contains(message, "evidence"):
		return CodeMissing
	case strings.Contains(message, "candidate"):
		return CodeCandidate
	default:
		return CodeMalformed
	}
}
func classifyBindingError(err error) issueClass {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "missing") || strings.Contains(message, "snapshot") {
		return issueUnknown
	}
	return issueFailClosed
}
func classifyBindingCode(err error) string {
	if strings.Contains(strings.ToLower(err.Error()), "snapshot") {
		return CodeStaleSnapshot
	}
	return CodeDigestMismatch
}
func bindEvidenceIDs(input Input) error {
	actual := make([]semantic.ID, 0, len(input.InferencePath.Evidence))
	for _, evidence := range input.InferencePath.Evidence {
		actual = append(actual, evidence.ID)
	}
	slices.Sort(actual)
	if equalIDs(actual, input.EvidenceIDs) {
		return nil
	}
	if len(actual) > len(input.EvidenceIDs) {
		return fmt.Errorf("missing evidence IDs")
	}
	return fmt.Errorf("orphan evidence IDs")
}
func bindSnapshots(input Input) error {
	for _, edge := range input.InferencePath.Edges {
		if edge.Before != input.Snapshots.Base || edge.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on edge %s", edge.RecordID)
		}
	}
	for _, claim := range input.InferencePath.Claims {
		if claim.Before != input.Snapshots.Base || claim.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on claim %s", claim.RecordID)
		}
	}
	for _, evidence := range input.InferencePath.Evidence {
		if evidence.Before != input.Snapshots.Base || evidence.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on evidence %s", evidence.ID)
		}
	}
	return nil
}
