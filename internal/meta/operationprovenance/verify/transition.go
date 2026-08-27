package verify

import "encoding/json"

func transitionFor(result metricResult, sourceDigest, semanticDigest string, f cFixture) claimTransition {
	artifacts := map[string]string{}
	for _, observation := range result.Relations {
		if observation.ObservedDigest != "" {
			artifacts[observation.Relation] = observation.ObservedDigest
		}
	}
	transition := claimTransition{Proposition: result.Proposition, Prior: result.Claim, Decision: result.Decision, SourceResolution: result.SourceResolution, LineageResolution: result.LineageResolution, Stage: "CLAIM", Provenance: provenance{Source: sourceDigest, Semantic: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, Operation: result.Lineage.Operation, Evidence: result.Lineage.Evidence, Mutation: f.mutation, Artifacts: artifacts}}
	switch result.Decision {
	case "PASS":
		transition.Next, transition.Transition, transition.Step, transition.Reason = "DISCHARGED", "DISCHARGED", "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		if result.Issue != nil && result.Issue.Cause == "VERIFIED_CONTRADICTION" {
			transition.Next, transition.Transition, transition.Step, transition.Reason = "REFUTED", "REFUTED", "refute-lineage-completeness-claim", "VERIFIED_CONTRADICTION_REFUTES_LINEAGE_CLAIM"
		} else {
			transition.Next, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNVERIFIED_FAILURE_PRESERVES_OPEN_CLAIM"
		}
	default:
		transition.Next, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		Proposition string                `json:"proposition"`
		MetricID    string                `json:"metric_id"`
		Decision    string                `json:"decision"`
		Issue       *issue                `json:"issue,omitempty"`
		Relations   []relationObservation `json:"relations"`
	}{result.Proposition, result.ID, result.Decision, result.Issue, result.Relations}
	payload, _ := json.Marshal(evidence)
	transition.EvidenceDigest = digest(payload)
	return transition
}
