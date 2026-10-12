package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

const compositionRecordHoldoutSource = `package staged
namespace staged
entity Integer id "staged://integer"
entity Box id "staged://box" fields { field value id "staged://box/value" type integer required one }
activity Seed(Integer) -> Box id "staged://seed" computes "return Box{value: input}" assembling {
 choice "increment" field_value at "0" alternative "input + 1" intent "Add one."
 value_case "[17]" -> "{\"value\":18}"
 holdout_value_case "[29]" -> "{\"value\":30}"
 attempts "2"
}
activity Wrap(Integer) -> Box id "staged://wrap" computes "let next = Seed(input); return Box{value: next.value * 2}" assembling {
 choice "double" field_value at "0" alternative "(next.value + 1) * 2" intent "Double the selected helper result."
 value_case "[17]" -> "{\"value\":36}"
 holdout_value_case "[29]" -> "{\"value\":60}"
 attempts "2"
}
activity Main(Integer) -> Integer id "staged://main" computes "let result = Wrap(input); return result.value"
`

func TestCompositionRecordHoldoutContinuationPreservesEachStage(t *testing.T) {
	source := []byte(compositionRecordHoldoutSource)
	ctx := context.Background()
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":41},"expected":{"Main":84}},{"inputs":{"Main":-9},"expected":{"Main":-16}}]}`)
	stop, policy := resumePolicyFixtures(t)
	prior, err := GenerateCompositionWithOptions(ctx, "staged.gooo", source, suite, CompositionOptions{EntryActivity: "Main", RecordPolicy: &stop})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(prior)
	if len(prior.Preparations) != 2 {
		t.Fatal("missing helper preparations")
	}
	for _, s := range prior.Preparations {
		r := s.Generation.Report.RecordAssembly
		if r.Holdout == nil || r.Holdout.Passed != 0 || r.Holdout.Total != 1 || len(r.Attempts) != 1 {
			t.Fatal(r)
		}
	}
	result, err := ResumeComposition(ctx, "staged.gooo", source, prior, suite, policy)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range result.Preparations {
		r := s.Generation.Report.RecordAssembly
		if r.Holdout == nil || r.Holdout.Passed != 1 || r.Holdout.Total != 1 || r.Passed != 1 || r.ModelCalls != 0 {
			t.Fatal(r)
		}
	}
	counts := result.Continuation.Activities
	if len(counts) != 2 || counts[0].AddedAttempts != 1 || counts[1].RecheckedAttempts != 1 || counts[1].AddedAttempts != 0 {
		t.Fatal(counts)
	}
	after, _ := json.Marshal(prior)
	if !bytes.Equal(before, after) {
		t.Fatal("continuation rewrote earlier held results")
	}
	encoded, _ := json.MarshalIndent(result, "", "  ")
	restored, err := DecodeComposition(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err = VerifyComposition(ctx, "staged.gooo", source, restored); err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "staged.gooo", source, restored, suite, nativeTool())
	if err != nil || native.FiniteTotal != 2 || native.FinitePassed != 2 || native.ModelCalls != 0 {
		t.Fatal(err, native)
	}
	restored.Preparations[1].Generation.Report.RecordAssembly.Holdout.Passed = 0
	if err = VerifyComposition(ctx, "staged.gooo", source, restored); err == nil {
		t.Fatal("changed final held score replayed")
	}
}
