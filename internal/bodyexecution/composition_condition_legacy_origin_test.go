package bodyexecution

import (
	"context"
	"testing"
)

func TestConditionOriginsLegacyAbsence2357(t *testing.T) {
	source := []byte(`package legacy2357
namespace legacy2357
entity Integer id "legacy2357://integer"
activity Shift(Integer) -> Integer id "legacy2357://shift" computes "if input <= 17 { return input * 2 }; return input - 9"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Shift":4},"expected":{"Shift":8}},{"inputs":{"Shift":23},"expected":{"Shift":14}}]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "legacy2357.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	for i := range prior.Steps {
		prior.Steps[i].Generation.Report.BodyConditions = nil
	}
	run, err := ExecuteComposition(ctx, "legacy2357.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || len(run.ConditionSites) != 1 {
		t.Fatal("fresh legacy absence execution differs", err, run.FinitePassed)
	}
	saveConditionObservation(t, "legacy2357", source, run)
	if run.ConditionSites[0].Source == nil || run.ConditionSites[0].Source.Reason != "NO_RETAINED_ORIGIN" || ConditionOriginFacts(run)[0].Selected {
		t.Fatal("legacy receipt invented a source origin", run.ConditionSites)
	}
	absent := run
	absent.ConditionSites = append([]CompositionConditionSite(nil), run.ConditionSites...)
	absent.ConditionSites[0].Source = nil
	if !sameJointRuntime(run, absent) {
		t.Fatal("explicit origin absence broke legacy comparison")
	}
	t.Log("legacy2357: native2/2; unavailable source origin retained as absence; no fabricated link")
}
