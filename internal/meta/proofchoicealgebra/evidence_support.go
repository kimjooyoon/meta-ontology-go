package proofchoicealgebra

import (
	"encoding/json"
	"sort"
	"strings"
)

func finishEvidence(result Evidence) Evidence {
	result.Producer = producerFor(result.Route)
	result.Consumer = "proof-choice-selector"
	result.ObservationIDs = slotIDs(result.ObservationSlots)
	result.EvidenceDigest = digestEvidence(result)
	return result
}

func producerFor(route Route) string {
	switch route {
	case Foundation:
		return "canonical-ir-foundation-producer"
	case Coherence:
		return "independent-projection-coherence-producer"
	case Regression:
		return "canonical-artifact-replay-producer"
	default:
		return ""
	}
}

func digestEvidence(result Evidence) string {
	result.EvidenceDigest = ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func evidenceFor(value Value, bundle evidenceBundle) []Evidence {
	return append([]Evidence(nil), bundle.ByValue[value.ID]...)
}

func slotIDs(slots []ObservationSlot) []string {
	result := make([]string, 0, len(slots))
	for _, slot := range slots {
		result = append(result, slot.ID)
	}
	sort.Strings(result)
	return result
}

func provenanceOf(slots []ObservationSlot) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, slot := range slots {
		for _, item := range slot.Provenance {
			if item != "" && !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
	}
	sort.Strings(result)
	return result
}

func stableSubject(subject string) bool {
	return strings.HasPrefix(subject, "gooo://") && !strings.Contains(subject, "unstable")
}
