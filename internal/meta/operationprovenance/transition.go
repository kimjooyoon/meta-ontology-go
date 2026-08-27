package operationprovenance

import "encoding/json"

func makeTransition(result MetricResult, sourceDigest, semanticDigest string, f fixture) ClaimTransition {
	artifacts := map[string]string{}
	for _, observation := range result.Relations {
		if observation.ObservedDigest != "" {
			artifacts[observation.Relation] = observation.ObservedDigest
		}
	}
	transition := ClaimTransition{Proposition: result.Proposition, PriorClaim: result.Claim, ConformanceDecision: result.Decision, SourceResolution: result.SourceResolution, LineageResolution: result.LineageResolution, Stage: "CLAIM", Provenance: Provenance{SourceDigest: sourceDigest, SemanticDigest: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, MetaOperation: result.Lineage.MetaOperation, EvidencePath: result.Lineage.EvidencePath, ScenarioMutation: f.Mutation, ArtifactDigests: artifacts}}
	switch result.Decision {
	case "PASS":
		transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "DISCHARGED", "DISCHARGED", "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		if result.Issue != nil && result.Issue.Cause == "VERIFIED_CONTRADICTION" {
			transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "REFUTED", "REFUTED", "refute-lineage-completeness-claim", "VERIFIED_CONTRADICTION_REFUTES_LINEAGE_CLAIM"
		} else {
			transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNVERIFIED_FAILURE_PRESERVES_OPEN_CLAIM"
		}
	default:
		transition.NextClaim, transition.Transition, transition.Step, transition.Reason = "OPEN", "PRESERVED_OPEN", "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		Proposition string                `json:"proposition"`
		MetricID    string                `json:"metric_id"`
		Decision    string                `json:"decision"`
		Issue       *Issue                `json:"issue,omitempty"`
		Relations   []RelationObservation `json:"relations"`
	}{result.Proposition, result.ID, result.Decision, result.Issue, result.Relations}
	payload, _ := json.Marshal(evidence)
	transition.EvidenceDigest = digestBytes(payload)
	return transition
}
