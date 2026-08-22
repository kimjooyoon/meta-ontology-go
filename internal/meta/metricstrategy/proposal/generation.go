package proposal

import (
	"bytes"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type generationFacts struct {
	Decision string
	Actions  int
	Writes   bool
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
	values := []struct{ ok bool; reason string; evidence any }{
		{actionable, "ACTIONABLE_GENERATION_PLAN_PROVEN", []any{first.Decision, first.Reason, first.PlanDigest, first.ReplayDigest, len(first.Selected)}},
		{independent, "INDEPENDENT_ACTION_GROUPS_PROVEN", first.Selected},
		{executable, "EXECUTABLE_CONFORMANCE_OBLIGATIONS_BOUND", first.Registry},
	}
	result := make([]Coordinate, 0, len(values))
	for offset, value := range values {
		status, reason := coordinateStatus(value.ok, false, value.reason)
		coordinate, err := makeCoordinate(offset+4, status, reason, value.evidence)
		if err != nil { return nil, generationFacts{}, err }
		result = append(result, coordinate)
	}
	return result, generationFacts{string(first.Decision), len(first.Selected), false, first.PromotionAuthorized || first.PromotionAuthorizedByPlan()}, nil
}

func conformancePlan() generation.Plan {
	indicator := func(metric sourcepolicy.Dimension, family sourcepolicy.Family, subject string, kind sourcepolicy.SubjectKind, value, limit int, proof sourcepolicy.ProofChoice, operation sourcepolicy.Operation) sourcepolicy.Indicator {
		return sourcepolicy.Indicator{MetricID: metric, Family: family, Subject: subject, SubjectKind: kind, Value: value, Limit: limit, Relation: sourcepolicy.RelationLessOrEqual, Applicability: sourcepolicy.ApplicabilityApplicable, ApplicabilityRule: sourcepolicy.ApplicabilityRuleDefault, ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable, Satisfied: false, Blocking: false, Proof: proof, Producer: "proposal-contract-conformance", Consumer: "generation.Build", Operation: operation}
	}
	report := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		indicator(sourcepolicy.DimensionRefactorAssign, sourcepolicy.FamilyRefactor, "fixture/expression.go:1:assignment", sourcepolicy.SubjectKindFunction, 2, 1, sourcepolicy.ProofRegression, sourcepolicy.OperationCollapseAssign),
		indicator(sourcepolicy.DimensionGoFileLines, sourcepolicy.FamilyVolume, "fixture/topology.go", sourcepolicy.SubjectKindFile, 76, 75, sourcepolicy.ProofFoundation, sourcepolicy.OperationSplitGo),
	}}
	return generation.Build(strings.Repeat("a", 40), strings.Repeat("b", 40), report)
}

func independentActions(actions []generation.Action) bool {
	if len(actions) != 2 { return false }
	return actions[0].IndependenceGroupID != actions[1].IndependenceGroupID && actions[0].IndicatorID != actions[1].IndicatorID && actions[0].Operation == sourcepolicy.OperationCollapseAssign && actions[1].Operation == sourcepolicy.OperationSplitGo
}

func executableActions(actions []generation.Action) bool {
	if len(actions) != 2 { return false }
	for _, action := range actions {
		if action.Executor == "" || action.Evaluator == "" || !action.ReceiptRequired || len(action.RequiredIndicatorIDs) == 0 || action.MetricProducer == "" || action.MetricConsumer == "" { return false }
	}
	return true
}
