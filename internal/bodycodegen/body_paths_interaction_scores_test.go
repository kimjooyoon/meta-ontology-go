package bodycodegen

import (
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

type exportedInteractionCases struct {
	cases      [][decision.DeclaredCaseFeatureDim]float32
	conditions [][decision.DeclaredConditionFeatureDim]float32
}

func (c exportedInteractionCases) CaseCount() int { return len(c.cases) }
func (c exportedInteractionCases) CaseFeatures(i int) ([decision.DeclaredCaseFeatureDim]float32, error) {
	return c.cases[i], nil
}
func (c exportedInteractionCases) ConditionCount() int { return len(c.conditions) }
func (c exportedInteractionCases) ConditionFeatureVersion() string {
	return decision.DeclaredConditionFeatureVersion
}
func (c exportedInteractionCases) ConditionFeatures(i int) ([decision.DeclaredConditionFeatureDim]float32, error) {
	return c.conditions[i], nil
}

func checkOrderedContractScores(t *testing.T, g *TypedPathGenerator,
	before TypedPathContextExport, rank *pathplan.ContractRanking) {
	t.Helper()
	inputs := make([][contractdecision.OrderedFeatureDim]float32, len(before.Inputs))
	for i, row := range before.Inputs {
		inputs[i] = *row.OrderedFeatures
	}
	cases := exportedInteractionCases{before.ContractCases.Features, before.ContractConditions.Features}
	var expected contractdecision.ChoicePrediction
	var err error
	switch model := g.condition.contract.(type) {
	case *contractdecision.InteractionRequirementModel:
		var workspace contractdecision.InteractionRequirementWorkspace
		err = model.PredictChoicesInto(inputs, cases, cases, &workspace, &expected)
	case *contractdecision.OrderedRequirementModel:
		var workspace contractdecision.OrderedRequirementWorkspace
		err = model.PredictChoicesInto(inputs, cases, cases, &workspace, &expected)
	default:
		t.Error("unexpected ordered contract model type")
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	if expected.Logits != rank.Logits || rank.ModelFingerprint != g.condition.contract.Fingerprint() {
		t.Error("actual compiler ranking differs from explicitly exported inputs")
	}
}
