package selfimprovementtermination

import "fmt"

func indicators(interventions []Intervention) []Indicator {
	return []Indicator{
		indicator(interventions[0], "gooo.termination.semantic-trace-intervention.v1", "SEMANTIC_TRACE_INTERVENTION", "SEMANTIC_TRACE_INTERVENTION_CHANGES_SEMANTIC_DIGEST"),
		indicator(interventions[1], "gooo.termination.nonsemantic-comment-intervention.v1", "NONSEMANTIC_COMMENT_INTERVENTION", "NONSEMANTIC_COMMENT_INTERVENTION_PRESERVES_SEMANTIC_DIGEST"),
	}
}

func indicator(intervention Intervention, id, route, reason string) Indicator {
	return Indicator{
		ID: id, Route: route, Producer: Producer, Consumer: Consumer,
		MetaOperation: MetaOperation, ProofChoice: ProofChoice, Stage: InterventionStage,
		Step: intervention.Step, Reason: reason,
		Value: fmt.Sprintf("source_changed=%t;semantic_changed=%t", intervention.SourceChanged, intervention.SemanticChanged),
		Limit: "source_changed=true;semantic_changed=true|false", Satisfied: true,
	}
}
