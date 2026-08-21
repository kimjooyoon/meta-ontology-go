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
		if indicator.Satisfied || indicator.Blocking {
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

func candidateKey(candidate candidate) string {
	return fmt.Sprintf("%020d\x00%s", candidate.binding.Priority, indicatorKey(candidate.indicator))
}

func actionFor(candidate candidate, id string) Action {
	return Action{IndicatorID: id, MetricID: candidate.indicator.MetricID, Subject: candidate.indicator.Subject,
		Operation: candidate.binding.Operation, IndependenceGroupID: candidate.binding.IndependenceGroupID,
		ProofChoice: candidate.binding.ProofChoice, Executor: candidate.binding.Executor, Priority: candidate.binding.Priority}
}
