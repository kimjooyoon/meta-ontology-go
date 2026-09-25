package semantic

import (
	"fmt"
	"sort"
)

func validateInferencePathRecords(out InferencePathV1, issues *InferencePathErrors) {
	seen := make(map[ID]string, len(out.Evidence)+len(out.Edges)+len(out.Claims))
	for _, record := range out.Evidence {
		registerInferenceID(seen, record.ID, "evidence", issues)
	}
	for _, edge := range out.Edges {
		registerInferenceID(seen, edge.RecordID, "edge", issues)
	}
	for _, claim := range out.Claims {
		registerInferenceID(seen, claim.RecordID, "claim", issues)
	}
	evidence := make(map[ID]InferenceEvidence, len(out.Evidence))
	for _, record := range out.Evidence {
		evidence[record.ID] = record
	}
	for _, edge := range out.Edges {
		validateInferenceEvidence(
			edge.InferenceRecord, edge.AcceptanceReceipt, edge.Kind == InferenceAcceptedLift, evidence, issues,
		)
		if edge.Kind == InferenceIndependentVerification && !hasIndependentEvidence(edge.Evidence, evidence) {
			issues.add("independent-evidence", edge.RecordID, "verification requires independent evidence")
		}
	}
	for _, claim := range out.Claims {
		validateInferenceEvidence(claim.InferenceRecord, "", false, evidence, issues)
	}
}
func sortInferenceIssues(issues *InferencePathErrors) {
	sort.Slice(issues.Issues, func(i, j int) bool {
		if issues.Issues[i].Code != issues.Issues[j].Code {
			return issues.Issues[i].Code < issues.Issues[j].Code
		}
		if issues.Issues[i].Record != issues.Issues[j].Record {
			return issues.Issues[i].Record < issues.Issues[j].Record
		}
		return issues.Issues[i].Detail < issues.Issues[j].Detail
	})
}
func registerInferenceID(seen map[ID]string, id ID, kind string, issues *InferencePathErrors) {
	if previous, exists := seen[id]; exists {
		issues.add("stable-id-collision", id, fmt.Sprintf("%s also used by %s", kind, previous))
		return
	}
	seen[id] = kind
}
