package verify

import "encoding/json"

func transitionFor(result metricResult, sourceDigest, semanticDigest string, f cFixture) claimTransition {
	transition := claimTransition{Prior: result.Claim, Decision: result.Decision, Resolution: result.Resolution, Stage: "CLAIM", Provenance: provenance{Source: sourceDigest, Semantic: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, Operation: result.Lineage.Operation, Evidence: result.Lineage.Evidence, Mutation: f.mutation}}
	switch result.Decision {
	case "PASS":
		transition.Next, transition.Transition, transition.Step, transition.Reason = "DISCHARGED", "DISCHARGED", "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		transition.Next, transition.Transition, transition.Step, transition.Reason = "REFUTED", "REFUTED", "refute-claim", "FAIL_CLOSED_IS_EXPLICIT_REFUTATION"
	default:
		transition.Next, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		MetricID string     `json:"metric_id"`
		Decision string     `json:"decision"`
		Issue    *issue     `json:"issue,omitempty"`
		Proof    provenance `json:"provenance"`
	}{result.ID, result.Decision, result.Issue, transition.Provenance}
	payload, _ := json.Marshal(evidence)
	transition.EvidenceDigest = digest(payload)
	return transition
}
