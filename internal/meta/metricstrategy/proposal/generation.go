package proposal

import (
	"bytes"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type generationFacts struct {
	Decision  string
	Actions   int
	Writes    bool
	Promotion bool
}

func generationCoordinates() ([]Coordinate, generationFacts, error) {
	first := conformancePlan()
	replay := conformancePlan()
	firstPayload, firstErr := generation.Encode(first)
	replayPayload, replayErr := generation.Encode(replay)
	actionable := firstErr == nil && replayErr == nil && bytes.Equal(firstPayload, replayPayload) && first.Decision == generation.DecisionPlan && first.Reason == generation.ReasonIndependentActions && len(first.Selected) == 2 && first.RequestedK == 2 && first.MinimumIndependent == 2
	independent := independentActions(first.Selected)
	executable := executableActions(first.Selected)
	values := []struct {
		ok       bool
		reason   string
		evidence any
	}{
		{actionable, "ACTIONABLE_GENERATION_PLAN_PROVEN", []any{first.Decision, first.Reason, first.PlanDigest, first.ReplayDigest, len(first.Selected)}},
		{independent, "INDEPENDENT_ACTION_GROUPS_PROVEN", first.Selected},
		{executable, "EXECUTABLE_CONFORMANCE_OBLIGATIONS_BOUND", first.Registry},
	}
	result := make([]Coordinate, 0, len(values))
	for offset, value := range values {
		status, reason := coordinateStatus(value.ok, false, value.reason)
		coordinate, err := makeCoordinate(offset+4, status, reason, value.evidence)
		if err != nil {
			return nil, generationFacts{}, err
		}
		result = append(result, coordinate)
	}
	return result, generationFacts{string(first.Decision), len(first.Selected), false, first.PromotionAuthorized || first.PromotionAuthorizedByPlan()}, nil
}
