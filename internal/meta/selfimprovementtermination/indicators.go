package selfimprovementtermination

import "fmt"

func indicators(interventions []Intervention) []Indicator {
	return []Indicator{
		indicator(interventions[0], "gooo.termination.semantic-trace-intervention.v1", "SEMANTIC_TRACE_INTERVENTION", "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST"),
		indicator(interventions[1], "gooo.termination.nonsemantic-comment-intervention.v1", "NONSEMANTIC_COMMENT_INTERVENTION", "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST"),
	}
}

func indicator(intervention Intervention, id, route, reason string) Indicator {
	satisfied := false
	limit := ""
	switch intervention.ID {
	case "semantic-trace":
		satisfied = intervention.SourceChanged && intervention.SemanticChanged &&
			(intervention.Baseline.Decision != intervention.Intervened.Decision ||
				intervention.Baseline.Resolution != intervention.Intervened.Resolution) &&
			canonicalOpenTransition(intervention.Intervened.ClaimTransitions)
		limit = "source_changed=true;semantic_changed=true;decision_or_resolution_changed;canonical_open_claim"
	case "nonsemantic-comment":
		satisfied = intervention.SourceChanged && !intervention.SemanticChanged &&
			intervention.SemanticBeforeDigest == intervention.SemanticAfterDigest &&
			sameOutcome(intervention.Baseline, intervention.Intervened)
		limit = "source_changed=true;semantic_changed=false;semantic_digest_same;outcome_same"
	}
	return Indicator{
		ID: id, Route: route, Producer: Producer, Consumer: Consumer,
		MetaOperation: MetaOperation, ProofChoice: ProofChoice, Stage: InterventionStage,
		Step: intervention.Step, Reason: reason,
		Value: fmt.Sprintf("baseline=%s/%s;intervened=%s/%s;claim=%s", intervention.Baseline.Decision,
			intervention.Baseline.Resolution, intervention.Intervened.Decision, intervention.Intervened.Resolution,
			intervention.Intervened.ClaimTransitions[len(intervention.Intervened.ClaimTransitions)-1].To),
		Limit: limit, Satisfied: satisfied,
	}
}

func countSatisfied(indicators []Indicator) int {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	return satisfied
}
