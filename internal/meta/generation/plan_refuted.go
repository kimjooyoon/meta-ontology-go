package generation

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func validatePlanRefutedEvidence(plan Plan) error {
	if !validOrderedIndicatorIDs(plan.RefutedIndicatorIDs) {
		return fmt.Errorf("refuted indicator IDs are not canonical")
	}
	seen := make(map[string]struct{}, len(plan.Counterexamples))
	seenIndicators := make(map[string]struct{}, len(plan.Counterexamples))
	refuted := make(map[string]struct{}, len(plan.RefutedIndicatorIDs))
	for _, indicatorID := range plan.RefutedIndicatorIDs {
		refuted[indicatorID] = struct{}{}
	}
	ledgerIndicators, err := refutedSourceIndicators(plan)
	if err != nil {
		return err
	}
	for _, counterexample := range plan.Counterexamples {
		if counterexample.ID == "" || counterexample.IndicatorID == "" ||
			counterexample.BlockerID == "" || counterexample.Stage == "" ||
			counterexample.Step == "" || counterexample.Reason == "" ||
			counterexample.UnknownClass == "" || counterexample.NextOperation == "" {
			return fmt.Errorf("incomplete planner counterexample")
		}
		if !validActionIndicatorID(counterexample.IndicatorID) {
			return fmt.Errorf("counterexample indicator ID is malformed")
		}
		if indicatorID(counterexample.SourceIndicator) != counterexample.IndicatorID ||
			counterexample.SourceIndicator.Satisfied ||
			counterexample.SourceIndicator.Applicability != sourcepolicy.ApplicabilityApplicable {
			return fmt.Errorf("counterexample source indicator is not the refuted observation")
		}
		if ledgerIndicators != nil {
			source, exists := ledgerIndicators[counterexample.IndicatorID]
			if !exists || !reflect.DeepEqual(source, counterexample.SourceIndicator) {
				return fmt.Errorf("counterexample source indicator does not match the ledger observation")
			}
		}
		binding, known := BindingForOperation(plan.Registry, counterexample.SourceIndicator.Operation)
		if counterexample.Reason == "INPUT_SUBJECT_KIND_MISMATCH" && !known {
			return fmt.Errorf("input-domain counterexample requires a known operation binding")
		}
		if known {
			if counterexample.SourceIndicator.SubjectKind == binding.InputSubjectKind {
				return fmt.Errorf("matching input-domain observation was marked refuted")
			}
			expected := inputDomainCounterexample(counterexample.SourceIndicator, binding)
			if !reflect.DeepEqual(counterexample, expected) {
				return fmt.Errorf("input-domain counterexample is not canonical")
			}
		}
		if _, exists := refuted[counterexample.IndicatorID]; !exists {
			return fmt.Errorf("counterexample is not linked to a refuted indicator")
		}
		if _, duplicate := seenIndicators[counterexample.IndicatorID]; duplicate {
			return fmt.Errorf("duplicate counterexample for refuted indicator %q", counterexample.IndicatorID)
		}
		if _, duplicate := seen[counterexample.ID]; duplicate {
			return fmt.Errorf("duplicate planner counterexample %q", counterexample.ID)
		}
		seen[counterexample.ID] = struct{}{}
		seenIndicators[counterexample.IndicatorID] = struct{}{}
	}
	if len(seenIndicators) != len(refuted) {
		return fmt.Errorf("refuted indicator is missing a planner counterexample")
	}
	return nil
}

func refutedSourceIndicators(plan Plan) (map[string]sourcepolicy.Indicator, error) {
	if plan.IndicatorDecisionLedger == nil {
		return nil, nil
	}
	result := make(map[string]sourcepolicy.Indicator, len(plan.IndicatorDecisionLedger.Entries))
	for _, entry := range plan.IndicatorDecisionLedger.Entries {
		if indicatorID(entry.SourceIndicator) != entry.IndicatorID {
			return nil, fmt.Errorf("ledger source indicator identity is not canonical")
		}
		if _, duplicate := result[entry.IndicatorID]; duplicate {
			return nil, fmt.Errorf("ledger source indicator is duplicated")
		}
		result[entry.IndicatorID] = entry.SourceIndicator
	}
	return result, nil
}
