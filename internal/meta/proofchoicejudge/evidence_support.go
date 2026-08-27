package proofchoicejudge

import (
	"encoding/json"
	"sort"
	"strings"
)

func finishEvidence(result evidence) evidence {
	result.Producer = producerFor(result.Route)
	result.Consumer = "proof-choice-selector"
	result.ObservationIDs = slotIDs(result.ObservationSlots)
	result.EvidenceDigest = digestEvidence(result)
	return result
}

func producerFor(route string) string {
	switch route {
	case "FOUNDATION":
		return "canonical-ir-foundation-producer"
	case "COHERENCE":
		return "independent-projection-coherence-producer"
	case "REGRESSION":
		return "canonical-artifact-replay-producer"
	default:
		return ""
	}
}

func digestEvidence(result evidence) string {
	result.EvidenceDigest = ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func evidenceFor(current value, bundle evidenceBundle) []evidence {
	return append([]evidence(nil), bundle.ByValue[current.ID]...)
}

func slotIDs(slots []observationSlot) []string {
	result := make([]string, 0, len(slots))
	for _, slot := range slots {
		result = append(result, slot.ID)
	}
	sort.Strings(result)
	return result
}

func provenanceOf(slots []observationSlot) []string {
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
