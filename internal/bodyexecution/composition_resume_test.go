package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func resumePolicyFixtures(t *testing.T) (bodycodegen.RecordAssemblyPolicy, bodycodegen.RecordAssemblyPolicy) {
	t.Helper()
	raw, err := os.ReadFile("../../examples/assembly-explainer/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	resume := bodycodegen.RecordAssemblyPolicy{Source: string(raw), Activity: "Explain"}
	stop := resume
	stop.Source = stop.Source[:strings.Index(stop.Source, "computes `")] + "computes `return Explanation{state: \"PROGRESS\", next_operation: \"USE_OBSERVED_CANDIDATE\", message: \"Checkpoint\"}`\n"
	return stop, resume
}

func resumeTwoRecordFixture(t *testing.T) (string, CompositionCases) {
	t.Helper()
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	activity := source[strings.Index(source, "activity Select"):strings.Index(source, "activity Label")]
	source += "\n" + strings.Replace(activity, "activity Select(", "activity SelectTwo(", 1)
	raw, err = os.ReadFile("../../examples/body-codegen/record-field-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	for i := range suite.Cases {
		row := &suite.Cases[i]
		row.Inputs["SelectTwo.input0"] = row.Inputs["Select.input0"]
		row.Inputs["SelectTwo.input1"] = row.Inputs["Select.input1"]
		row.Expected["SelectTwo"] = row.Expected["Select"]
	}
	return source, suite
}

func TestCompositionResumesMultipleRecordActivitiesAndNativeReplay(t *testing.T) {
	ctx := context.Background()
	source, suite := resumeTwoRecordFixture(t)
	stop, policy := resumePolicyFixtures(t)
	prior, err := GenerateCompositionWithOptions(ctx, "two.gooo", []byte(source), suite, CompositionOptions{RecordPolicy: &stop})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(prior)
	result, err := ResumeComposition(ctx, "two.gooo", []byte(source), prior, suite, policy)
	if err != nil {
		t.Fatal(err)
	}
	if result.Model != nil || result.Continuation.NewModelCalls != 0 || len(result.Continuation.Activities) != 2 {
		t.Fatal("continued graph lost activity accounting", result.Continuation)
	}
	for _, counts := range result.Continuation.Activities {
		if counts.RetainedAttempts != 1 || counts.AddedAttempts != 7 {
			t.Fatal("graph restarted an activity", counts)
		}
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("graph resume mutated predecessor")
	}
	encoded, _ := json.Marshal(result)
	restored, err := DecodeComposition(encoded)
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "two.gooo", []byte(source), restored, suite, nativeTool())
	if err != nil || native.FinitePassed != 21 || native.ModelCalls != 0 || !native.ProjectionReplayed {
		t.Fatal("continued graph did not execute/replay", err, native)
	}
	result.Continuation.Activities[0].AddedAttempts++
	if err := VerifyComposition(ctx, "two.gooo", []byte(source), result); err == nil {
		t.Fatal("changed continuation counts replayed")
	}
}
