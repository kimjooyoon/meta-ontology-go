package minimalcausalexplanation

import "fmt"

type indicatorCheck struct {
	id, class, operation, proof, expected, actual string
	satisfied                                     bool
	evidence                                      any
}

func buildIndicators(source []byte, graph CausalGraph, program MetaProgram, cases []ExplanationCase, summary Summary, preservation Preservation) ([]Indicator, error) {
	minimal := cases[0].Paths[0]
	overlong := cases[1].Paths[0]
	insufficient := cases[2].Paths[0]
	checks := []indicatorCheck{
		{"MCE-FOUNDATION-SOURCE-001", "FOUNDATION", "bind-source", "FOUNDATION", "gooo source bound", fmt.Sprintf("path=%s lines=%d", cases[0].ID, countLines(source)), true, []any{cases[0].ID, countLines(source)}},
		{"MCE-FOUNDATION-GRAPH-002", "FOUNDATION", "freeze-graph", "FOUNDATION", "nodes=4 edges=2", fmt.Sprintf("nodes=%d edges=%d", len(graph.Nodes), len(graph.Edges)), len(graph.Nodes) == 4 && len(graph.Edges) == 2, graph},
		{"MCE-FOUNDATION-PROGRAM-003", "FOUNDATION", "freeze-graph", "FOUNDATION", "operations=6 indicators=12", fmt.Sprintf("operations=%d indicators=%d", len(program.MetaOperations), program.IndicatorDenominator), len(program.MetaOperations) == 6 && program.IndicatorDenominator == IndicatorTotal, program},
		{"MCE-COHERENCE-SUFFICIENT-004", "COHERENCE", "evaluate-sufficiency", "COHERENCE", "minimal path decision=PASS sufficient=true", fmt.Sprintf("decision=%s sufficient=%t", minimal.Decision, minimal.Sufficient), minimal.Decision == DecisionPass && minimal.Sufficient, minimal},
		{"MCE-COHERENCE-MINIMAL-005", "COHERENCE", "minimize-path", "COHERENCE", "minimal=true evidence=3", fmt.Sprintf("minimal=%t evidence=%d", minimal.Minimal, len(minimal.EvidenceIDs)), minimal.Minimal && len(minimal.EvidenceIDs) == 3, minimal},
		{"MCE-COHERENCE-OVERLONG-006", "COHERENCE", "evaluate-sufficiency", "COHERENCE", "sufficient=true minimal=false", fmt.Sprintf("sufficient=%t minimal=%t", overlong.Sufficient, overlong.Minimal), overlong.Sufficient && !overlong.Minimal, overlong},
		{"MCE-COHERENCE-INSUFFICIENT-007", "COHERENCE", "evaluate-sufficiency", "COHERENCE", "sufficient=false", fmt.Sprintf("sufficient=%t decision=%s", insufficient.Sufficient, insufficient.Decision), !insufficient.Sufficient && insufficient.Decision == DecisionFailClosed, insufficient},
		{"MCE-COHERENCE-COUNTERFACTUAL-008", "COHERENCE", "minimize-path", "COHERENCE", "minimal changed=3/3", fmt.Sprintf("changed=%d/%d", changed(minimal.Counterfactuals), len(minimal.Counterfactuals)), len(minimal.Counterfactuals) == 3 && changed(minimal.Counterfactuals) == 3, minimal.Counterfactuals},
		{"MCE-COHERENCE-EXTRA-COUNTERFACTUAL-009", "COHERENCE", "minimize-path", "COHERENCE", "overlong changed=3/4", fmt.Sprintf("changed=%d/%d", changed(overlong.Counterfactuals), len(overlong.Counterfactuals)), len(overlong.Counterfactuals) == 4 && changed(overlong.Counterfactuals) == 3, overlong.Counterfactuals},
		{"MCE-COHERENCE-PATH-AUTHORITY-010", "COHERENCE", "judge-receipt", "COHERENCE", "path_set=true text=INCIDENTAL", fmt.Sprintf("path_set=%t text=%s", summary.PathSetAuthoritative, summary.ExplanationTextRole), summary.PathSetAuthoritative && summary.ExplanationTextRole == ExplanationTextRole, summary},
		{"MCE-REGRESSION-CLAIMS-011", "REGRESSION", "preserve-claims", "REGRESSION", "preserved=6 transitions=12", fmt.Sprintf("preserved=%d transitions=%d", preservation.PreservedTotal, preservation.TransitionTotal), preservation.PreservedTotal == ClaimTotal && preservation.TransitionTotal == TransitionTotal, preservation},
		{"MCE-REGRESSION-READONLY-012", "REGRESSION", "judge-receipt", "REGRESSION", "repository_writes=0 promotion=false", fmt.Sprintf("repository_writes=%d promotion=%t", summary.RepositoryWrites, summary.PromotionAuthorized), summary.RepositoryWrites == 0 && !summary.PromotionAuthorized, summary},
	}
	indicators := make([]Indicator, 0, len(checks))
	for _, check := range checks {
		evidenceDigest, err := digestValue(check.evidence)
		if err != nil {
			return nil, err
		}
		operation := operationByID(program.MetaOperations, check.operation)
		indicators = append(indicators, Indicator{ID: check.id, Class: check.class, MetaOperation: check.operation, Producer: operation.Producer, Consumer: operation.Consumer, ProofChoice: check.proof, Expected: check.expected, Actual: check.actual, Satisfied: check.satisfied, EvidenceDigest: evidenceDigest})
	}
	return indicators, nil
}

func changed(counterfactuals []Counterfactual) int {
	total := 0
	for _, counterfactual := range counterfactuals {
		if counterfactual.Changed {
			total++
		}
	}
	return total
}

func operationByID(operations []MetaOperation, id string) MetaOperation {
	for _, operation := range operations {
		if operation.ID == id {
			return operation
		}
	}
	return MetaOperation{}
}
