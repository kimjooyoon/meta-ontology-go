package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func pureCallCompositionFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/pure-activity-calls/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/assembly-policy/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionPureCallsNativeAssemblyAndReplay(t *testing.T) {
	ctx := context.Background()
	source, suite := pureCallCompositionFixture(t)
	prior, err := GenerateCompositionWithOptions(ctx, "calls.gooo", source, suite, CompositionOptions{EntryActivity: "Diagnose"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Plan.Activities) != 1 || prior.Plan.EntryActivity != "Diagnose" ||
		prior.Steps[0].Generation.Report.CallClosure == nil || !strings.Contains(prior.Source, "GoooComposedActivity0Call0") {
		t.Fatal("helper became a separate input activity", prior.Plan)
	}
	native, err := ExecuteComposition(ctx, "calls.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 4 || native.FiniteTotal != 4 || !native.RuntimeReplayed {
		t.Fatal("native call result", err, native)
	}
	encoded, _ := json.Marshal(prior)
	restored, err := DecodeComposition(encoded)
	if err != nil || VerifyComposition(ctx, "calls.gooo", source, restored) != nil {
		t.Fatal("call composition did not replay", err)
	}
	prior.Steps[0].Generation.Report.CallClosure.Activities[0].ProgramSHA256 = "changed"
	if err := VerifyComposition(ctx, "calls.gooo", source, prior); err == nil {
		t.Fatal("changed callee identity replayed")
	}
}

func TestCompositionPureCallsRetainEntryThroughContinuation(t *testing.T) {
	ctx := context.Background()
	source, suite := pureCallCompositionFixture(t)
	stop, policy := resumePolicyFixtures(t)
	prior, err := GenerateCompositionWithOptions(ctx, "calls.gooo", source, suite,
		CompositionOptions{EntryActivity: "Diagnose", RecordPolicy: &stop})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ResumeComposition(ctx, "calls.gooo", source, prior, suite, policy)
	if err != nil || result.Plan.EntryActivity != "Diagnose" || result.Continuation.Activities[0].AddedAttempts != 3 {
		t.Fatal("entry continuation", err)
	}
	if _, err := GenerateCompositionWithOptions(ctx, "calls.gooo", source, suite, CompositionOptions{EntryActivity: "Missing"}); err == nil {
		t.Fatal("unknown entry accepted")
	}
}

func TestCompositionEntryKeepsExplicitBindProducers(t *testing.T) {
	source := []byte("package calls\nnamespace calls\nentity Integer id \"gooo://calls/integer\"\n" +
		"activity Input(Integer) -> Integer computes `return input + 1`\n" +
		"activity Twice(Integer) -> Integer computes `return input * 2`\n" +
		"activity Score(Integer) -> Integer computes `return Twice(input) + 3`\n" +
		"bind Input.result -> Score.input\n")
	suite, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Input":4},"expected":{"Input":5,"Score":13}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "calls.gooo", source, suite, CompositionOptions{EntryActivity: "Score"})
	if err != nil || len(prior.Plan.Activities) != 2 || len(prior.Plan.Edges) != 1 {
		t.Fatal("entry lost a typed bind producer", err, prior.Plan)
	}
	native, err := ExecuteComposition(ctx, "calls.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 2 || native.FiniteTotal != 2 {
		t.Fatal("entry bind and call projection", err, native)
	}
}
