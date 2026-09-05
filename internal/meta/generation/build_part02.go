package generation

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type candidate struct {
	indicator sourcepolicy.Indicator
	binding   Binding
}

func partitionIndicators(indicators []sourcepolicy.Indicator) ([]sourcepolicy.Indicator, []string) {
	actionable, unknown := make([]sourcepolicy.Indicator, 0), make([]string, 0)
	for _, indicator := range indicators {
		if indicator.Applicability == sourcepolicy.ApplicabilityNotApplicable ||
			indicator.Satisfied {
			continue
		}
		if sourcepolicy.IsLineCapMetric(indicator.MetricID) && !indicator.IsDriverCandidate() {
			unknown = append(unknown, indicatorID(indicator))
			continue
		}
		if indicator.Operation == "" {
			unknown = append(unknown, indicatorID(indicator))
			continue
		}
		actionable = append(actionable, indicator)
	}
	return actionable, unknown
}

func partitionIndicatorsForRegistry(indicators []sourcepolicy.Indicator, registry []Binding) ([]sourcepolicy.Indicator, []string, []string, []Counterexample) {
	actionable, unknown := partitionIndicators(indicators)
	index, valid := registryIndex(registry)
	if !valid {
		return actionable, unknown, nil, nil
	}
	routable := make([]sourcepolicy.Indicator, 0, len(actionable))
	refuted := make([]string, 0)
	counterexamples := make([]Counterexample, 0)
	for _, indicator := range actionable {
		binding, exists := index[indicator.Operation]
		if !exists || indicator.SubjectKind == binding.InputSubjectKind {
			routable = append(routable, indicator)
			continue
		}
		id := indicatorID(indicator)
		refuted = append(refuted, id)
		counterexamples = append(counterexamples, inputDomainCounterexample(indicator, binding))
	}
	return routable, unknown, refuted, counterexamples
}

func inputDomainCounterexample(indicator sourcepolicy.Indicator, binding Binding) Counterexample {
	return Counterexample{
		ID:            "input-domain:" + indicatorID(indicator),
		IndicatorID:   indicatorID(indicator),
		SourceIndicator: indicator,
		BlockerID:     "binding-input-domain:" + string(binding.Operation) + ":" + string(binding.InputSubjectKind) + ":" + string(indicator.SubjectKind),
		Stage:         "binding",
		Step:          "validate-input-subject-kind",
		Reason:        "INPUT_SUBJECT_KIND_MISMATCH",
		UnknownClass:  "KNOWN_CONTRADICTION",
		NextOperation: "select-valid-domain-action",
		BlockedBy:     []string{string(binding.Operation)},
	}
}

func selectActions(plan Plan, indicators []sourcepolicy.Indicator, registry []Binding) Plan {
	index, _ := registryIndex(registry)
	candidates := make([]candidate, 0, len(indicators))
	for _, indicator := range indicators {
		binding, exists := index[indicator.Operation]
		if !exists {
			plan.UnknownIndicatorIDs = append(plan.UnknownIndicatorIDs, indicatorID(indicator))
			continue
		}
		candidates = append(candidates, candidate{indicator: indicator, binding: binding})
	}
	if len(plan.UnknownIndicatorIDs) != 0 {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonMissingOperation
		return finish(plan)
	}
	if len(candidates) == 0 {
		if len(plan.RefutedIndicatorIDs) != 0 {
			plan.Shortfall = []string{fmt.Sprintf("independent-groups:0/%d", minimumIndependent)}
			plan.Decision, plan.Reason = DecisionUnknown, ReasonPressureShortfall
			return finish(plan)
		}
		plan.Decision, plan.Reason = DecisionFixedPoint, ReasonExactFixedPoint
		return finish(plan)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidateKey(candidates[i]) < candidateKey(candidates[j]) })
	groups := make(map[string]struct{})
	for _, candidate := range candidates {
		id := indicatorID(candidate.indicator)
		if _, used := groups[candidate.binding.IndependenceGroupID]; used || uint32(len(plan.Selected)) == requestedK {
			plan.UnselectedIndicatorIDs = append(plan.UnselectedIndicatorIDs, id)
			continue
		}
		groups[candidate.binding.IndependenceGroupID] = struct{}{}
		plan.Selected = append(plan.Selected, actionFor(candidate, id))
	}
	if uint32(len(plan.Selected)) < minimumIndependent {
		for _, action := range plan.Selected {
			plan.UnselectedIndicatorIDs = append(plan.UnselectedIndicatorIDs, action.IndicatorID)
		}
		plan.Selected = []Action{}
		plan.Shortfall = []string{fmt.Sprintf("independent-groups:%d/%d", len(groups), minimumIndependent)}
		plan.Decision, plan.Reason = DecisionUnknown, ReasonPressureShortfall
		return finish(plan)
	}
	plan.Decision, plan.Reason = DecisionPlan, ReasonIndependentActions
	return finish(plan)
}
