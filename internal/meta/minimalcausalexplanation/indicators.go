package minimalcausalexplanation

import "fmt"

type indicatorCheck struct {
	id, class, operation, proof, expected, actual string
	satisfied                                     bool
	evidence                                      any
}

func buildIndicators(model sourceModel, result assessment, boundary RepositoryBoundary, preservation Preservation, regression ClaimRegression, interventions []Intervention) ([]Indicator, error) {
	minimal := result.Cases[0].Paths[0]
	overlong := result.Cases[1].Paths[0]
	insufficient := result.Cases[2].Paths[0]
	checks := []indicatorCheck{
		{"MCE-FOUNDATION-SOURCE-001", "FOUNDATION", "bind-source", "FOUNDATION", "ast=true ir=true", fmt.Sprintf("ast=%t ir=%t", model.SourceReconstruct.ASTParsed, model.SourceReconstruct.IRLowered), model.SourceReconstruct.ASTParsed && model.SourceReconstruct.IRLowered, model.SourceReconstruct},
		{"MCE-FOUNDATION-GRAPH-002", "FOUNDATION", "bind-source", "FOUNDATION", "source graph reconstructed", fmt.Sprintf("nodes=%d edges=%d", len(model.Graph.Nodes), len(model.Graph.Edges)), model.SourceReconstruct.GraphReconstructed, model.Graph},
		{"MCE-FOUNDATION-PREDICATE-003", "FOUNDATION", "judge-predicate", "FOUNDATION", "predicate reconstructed from gooo value", model.Predicate.Value, model.SourceReconstruct.PredicateReconstructed && model.Predicate.Value != "", model.Predicate},
		{"MCE-FOUNDATION-PROGRAM-004", "FOUNDATION", "preserve-claims", "FOUNDATION", "operations=6 indicators=12", fmt.Sprintf("operations=%d indicators=%d", len(model.Program.MetaOperations), model.Program.IndicatorDenominator), len(model.Program.MetaOperations) == 6 && model.Program.IndicatorDenominator == 12, model.Program},
		{"MCE-COHERENCE-OBSERVED-005", "COHERENCE", "bind-compiler-receipt", "COHERENCE", "observed evidence=3", fmt.Sprintf("observed evidence=%d", result.Summary.ObservedEvidenceTotal), result.Summary.ObservedEvidenceTotal == 3, result.Observed},
		{"MCE-COHERENCE-SUBSET-006", "COHERENCE", "judge-predicate", "COHERENCE", "subset-minimal=1/2", fmt.Sprintf("subset-minimal=%d/%d", result.Summary.SubsetMinimalNumerator, result.Summary.SubsetMinimalDenominator), result.Summary.SubsetMinimalNumerator == 1 && result.Summary.SubsetMinimalDenominator == 2, minimal},
		{"MCE-COHERENCE-CARDINALITY-007", "COHERENCE", "judge-predicate", "COHERENCE", "cardinality-minimum=1/2", fmt.Sprintf("cardinality-minimum=%d/%d", result.Summary.CardinalityMinimumNumerator, result.Summary.CardinalityMinimumDenominator), result.Summary.CardinalityMinimumNumerator == 1 && result.Summary.CardinalityMinimumDenominator == 2 && minimal.CombinationSearch.Exhaustive, minimal.CombinationSearch},
		{"MCE-COHERENCE-OVERLONG-008", "COHERENCE", "judge-predicate", "COHERENCE", "overlong subset/cardinality rejected", fmt.Sprintf("subset=%s cardinality=%s", overlong.SubsetMinimal, overlong.CardinalityMinimum), overlong.Sufficient && overlong.SubsetMinimal == NotSubsetMinimal && overlong.CardinalityMinimum == NotCardinalityMinimum, overlong},
		{"MCE-COHERENCE-COUNTERFACTUAL-009", "COHERENCE", "judge-predicate", "COHERENCE", "changed=6/7 executions", fmt.Sprintf("changed=%d executions=%d", result.Summary.ChangedCounterfactuals, result.Summary.CounterfactualExecutions), result.Summary.ChangedCounterfactuals == 6 && result.Summary.CounterfactualExecutions == 7, result.Cases},
		{"MCE-REGRESSION-CLAIMS-010", "REGRESSION", "preserve-claims", "REGRESSION", "failure regression refutes instead of discharging", fmt.Sprintf("legacy=%s correct=%s", regression.LegacyUnconditionalState, regression.CorrectState), regression.LegacyUnconditionalState == ClaimDischarged && regression.CorrectState == ClaimRefuted && hasRefutedTransition(regression.Transitions), regression},
		{"MCE-REGRESSION-INTERVENTION-011", "REGRESSION", "judge-predicate", "REGRESSION", "2 semantic changes and 1 comment preservation", fmt.Sprintf("interventions=%d", len(interventions)), interventionContract(interventions), interventions},
		{"MCE-REGRESSION-READONLY-012", "REGRESSION", "preserve-claims", "REGRESSION", "before/after writes=0 promotion=false", fmt.Sprintf("writes=%d promotion=%t", boundary.Writes, boundary.PromotionAuthorized), boundary.Writes == 0 && !boundary.PromotionAuthorized && preservation.PreservedTotal == preservation.ClaimTotal, boundary},
	}
	indicators := make([]Indicator, 0, len(checks))
	for _, check := range checks {
		op := operationByID(model.Program.MetaOperations, check.operation)
		digest, err := digestValue(check.evidence)
		if err != nil {
			return nil, err
		}
		indicators = append(indicators, Indicator{ID: check.id, Class: check.class, MetaOperation: check.operation, Producer: op.Producer, Consumer: op.Consumer, ProofChoice: check.proof, Expected: check.expected, Actual: check.actual, Satisfied: check.satisfied && op.ID != "", EvidenceDigest: digest})
	}
	return indicators, nil
}

func operationByID(operations []MetaOperation, id string) MetaOperation {
	for _, operation := range operations {
		if operation.ID == id {
			return operation
		}
	}
	return MetaOperation{}
}

func hasRefutedTransition(transitions []ClaimTransition) bool {
	for _, transition := range transitions {
		if transition.After == ClaimRefuted {
			return true
		}
	}
	return false
}

func interventionContract(interventions []Intervention) bool {
	if len(interventions) != 3 {
		return false
	}
	return interventions[0].SemanticChanged && interventions[0].AfterDecision == StatusFailClosed && interventions[0].PathSetChanged &&
		interventions[1].SemanticChanged && interventions[1].AfterDecision == StatusFailClosed && interventions[1].PathSetChanged &&
		!interventions[2].SemanticChanged && interventions[2].SemanticDigestPreserved && interventions[2].ResultPreserved
}
