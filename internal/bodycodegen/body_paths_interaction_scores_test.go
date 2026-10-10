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

func checkInteractionContractScores(t *testing.T, g *TypedPathGenerator,
	before TypedPathContextExport, rank *pathplan.ContractRanking) {
	t.Helper()
	inputs := make([][contractdecision.OrderedFeatureDim]float32, len(before.Inputs))
	for i, row := range before.Inputs {
		inputs[i] = *row.OrderedFeatures
	}
	cases := exportedInteractionCases{before.ContractCases.Features, before.ContractConditions.Features}
	var workspace contractdecision.InteractionRequirementWorkspace
	var expected contractdecision.ChoicePrediction
	model := g.condition.contract.(*contractdecision.InteractionRequirementModel)
	if err := model.PredictChoicesInto(inputs, cases, cases, &workspace, &expected); err != nil {
		t.Error(err)
		return
	}
	if expected.Logits != rank.Logits || rank.ModelFingerprint != model.Fingerprint() {
		t.Error("actual compiler ranking differs from explicitly exported inputs")
	}
}
