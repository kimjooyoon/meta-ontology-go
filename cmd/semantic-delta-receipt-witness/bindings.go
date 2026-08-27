package main

import producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/semanticdeltareceipt"

func bindings(summary Summary) []producer.OperationBinding {
	return []producer.OperationBinding{
		binding(producer.MetricID, "OUTCOME", "COHERENCE", "cases", "GREATER_OR_EQUAL", summary.CasesPassed, summary.CasesTotal, summary.CasesTotal, "reduce", "case-suite", "ALL_FIXED_DENOMINATOR_CASES_REPLAYED"),
		binding("gooo.metric.evidence.textual-delta-observation.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.TextualChanges, summary.CasesTotal, summary.CasesTotal, "observe", "raw-bytes", "TEXTUAL_BYTES_BOUND"),
		binding("gooo.metric.evidence.structural-delta-separation.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.StructuralObservations, summary.CasesTotal, summary.CasesTotal, "derive", "canonical-graph", "STRUCTURAL_GRAPH_BOUND_SEPARATELY"),
		binding("gooo.metric.semantic.claim-transition-totality.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.ClaimTransitionCases, summary.CasesTotal, summary.CasesTotal, "derive", "claim-transition", "CLAIM_TRANSITIONS_EXPLICIT"),
		binding("gooo.metric.epistemic.delta-receipt-adjudication.v1", "GUARDRAIL", "COHERENCE", "cases", "GREATER_OR_EQUAL", summary.AdjudicatedCases, summary.CasesTotal, summary.CasesTotal, "adjudicate", "independent-replay", "INDEPENDENT_JUDGE_REPLAYED"),
		binding("gooo.metric.effects.delta-receipt-writes.v1", "GUARDRAIL", "REGRESSION", "writes", "LESS_OR_EQUAL", summary.RepositoryWrites, 0, summary.CasesTotal, "observe", "read-only-boundary", "NO_REPOSITORY_WRITES"),
	}
}

func binding(metricID, class, proof, unit, relation string, value, target, denominator int, stage, step, reason string) producer.OperationBinding {
	satisfied := denominator > 0 && value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = denominator > 0 && value <= target
	}
	return producer.OperationBinding{MetricID: metricID, Class: class, ProofChoice: proof, Producer: producer.Producer, Consumer: producer.Consumer, MetaOperation: producer.MetaOperation, Stage: stage, Step: step, Reason: reason, Unit: unit, Relation: relation, Value: value, Target: target, Denominator: denominator, Satisfied: satisfied}
}
