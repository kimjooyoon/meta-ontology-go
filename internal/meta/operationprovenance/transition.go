package operationprovenance

import "encoding/json"

func makeTransition(result MetricResult, sourceDigest, semanticDigest string, f fixture) ClaimTransition {
	transition := ClaimTransition{PriorClaim: result.Claim, ConformanceDecision: result.Decision, SubjectResolution: result.SubjectResolution, Stage: "CLAIM", Provenance: Provenance{SourceDigest: sourceDigest, SemanticDigest: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, MetaOperation: result.Lineage.MetaOperation, EvidencePath: result.Lineage.EvidencePath, ScenarioMutation: f.Mutation}}
	switch result.Decision {
	case "PASS":
		transition.NextClaim, transition.Transition = "DISCHARGED", "DISCHARGED"
		transition.Step, transition.Reason = "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		transition.NextClaim, transition.Transition = "REFUTED", "REFUTED"
		transition.Step, transition.Reason = "refute-claim", "FAIL_CLOSED_IS_EXPLICIT_REFUTATION"
	default:
		transition.NextClaim, transition.Transition = "OPEN", "PRESERVED_OPEN"
		transition.Step, transition.Reason = "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		MetricID string     `json:"metric_id"`
		Decision string     `json:"decision"`
		Issue    *Issue     `json:"issue,omitempty"`
		Proof    Provenance `json:"provenance"`
	}{result.ID, result.Decision, result.Issue, transition.Provenance}
	payload, _ := json.Marshal(evidence)
	transition.EvidenceDigest = digestBytes(payload)
	return transition
}
