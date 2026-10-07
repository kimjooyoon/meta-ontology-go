package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCompositionResumesHelpersAndInvalidatesCompleteCaller(t *testing.T) {
	source := []byte(calledRecordSeed + `
activity Wrap(Integer) -> Box computes "let next = Seed(input); return Box{value: next.value * 2}" assembling {
    choice "value" field_value at "0" alternative "(next.value - 1) * 2" intent "Double the original input."
    value_case "[1]" -> "{\"value\":2}"
    value_case "[2]" -> "{\"value\":4}"
    attempts "2"
}
activity Main(Integer) -> Integer computes "let left = Wrap(input); let right = Wrap(input + 1); return left.value + right.value"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":3},"expected":{"Main":14}}]}`)
	ctx := context.Background()
	stop, policy := resumePolicyFixtures(t)
	prior, err := GenerateCompositionWithOptions(ctx, "nested.gooo", source, suite, CompositionOptions{EntryActivity: "Main", RecordPolicy: &stop})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(prior)
	if prior.Preparations[1].Generation.Report.RecordAssembly.Passed != 2 {
		t.Fatal("caller was not initially complete")
	}
	result, err := ResumeComposition(ctx, "nested.gooo", source, prior, suite, policy)
	if err != nil {
		t.Fatal(err)
	}
	counts := result.Continuation.Activities
	if len(counts) != 2 || counts[0].Activity != "Seed" || counts[0].RecheckedAttempts != 0 || counts[0].AddedAttempts != 1 ||
		counts[1].Activity != "Wrap" || counts[1].RecheckedAttempts != 1 || counts[1].AddedAttempts != 1 {
		t.Fatal(counts)
	}
	caller := result.Preparations[1].Generation.Report.RecordAssembly
	if caller.Attempts[0].Passed != 0 || caller.Passed != 2 || caller.SelectedMask != 1 {
		t.Fatal(caller)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("prior composition mutated")
	}
	native, err := ExecuteComposition(ctx, "nested.gooo", source, result, suite, nativeTool())
	if err != nil || native.FinitePassed != 1 || native.ModelCalls != 0 || !native.ProjectionReplayed {
		t.Fatal(err, native)
	}
	// Repeated continuation must preserve the historical source transition while
	// adding neither attempts nor dependency rechecks to the current round.
	result, err = ResumeComposition(ctx, "nested.gooo", source, result, suite, policy)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range result.Continuation.Activities {
		if c.RecheckedAttempts != 0 || c.AddedAttempts != 0 {
			t.Fatal(c)
		}
	}
	if err := VerifyComposition(ctx, "nested.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	// Even a self-consistent standalone history cannot claim a helper checkpoint
	// that this composition did not construct in that round.
	result.Preparations[1].Generation.Report.RecordAssembly.ControlHistory[0].SourceSHA256 = digest([]byte("invented helper"))
	if err := verifyCalledContinuationHistory(ctx, "nested.gooo", source, result); err == nil {
		t.Fatal("foreign checkpoint accepted")
	}
}

func TestCompositionResumesBoundAndCalledActivityOnlyOnce(t *testing.T) {
	source := []byte(calledRecordSeed + "activity Main(Box) -> Box computes `return Seed(input.value)`\nbind Seed.result -> Main.input\n")
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Seed":2},"expected":{"Seed":{"value":3},"Main":{"value":4}}}]}`)
	ctx := context.Background()
	stop, policy := resumePolicyFixtures(t)
	prior, err := GenerateCompositionWithOptions(ctx, "bound.gooo", source, suite, CompositionOptions{EntryActivity: "Main", RecordPolicy: &stop})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ResumeComposition(ctx, "bound.gooo", source, prior, suite, policy)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Continuation.Activities) != 1 || result.Steps[0].Generation.Report.RecordAssembly != nil {
		t.Fatal("prepared activity counted twice")
	}
	native, err := ExecuteComposition(ctx, "bound.gooo", source, result, suite, nativeTool())
	if err != nil || native.FinitePassed != 2 {
		t.Fatal(err, native)
	}
}
