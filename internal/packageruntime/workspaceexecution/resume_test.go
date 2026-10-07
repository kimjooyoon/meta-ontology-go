package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func continuationManifest(t *testing.T, example, entry string) packageruntime.Manifest {
	t.Helper()
	raw, err := os.ReadFile("../../../examples/" + example)
	if err != nil {
		t.Fatal(err)
	}
	m := standaloneWorkspace(string(raw), entry)
	m.Packages[0].Name = strings.Fields(string(raw))[1]
	return m
}

func continuationFixture(t *testing.T) (packageruntime.Manifest, bodyexecution.CompositionCases, packageruntime.Manifest, packageruntime.Manifest) {
	t.Helper()
	m := continuationManifest(t, "dependent-continuation/main.gooo.fixture", "Main")
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/tool:Main":3},"expected":{"example/tool:Main":18}},
{"inputs":{"example/tool:Main":-2},"expected":{"example/tool:Main":-2}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	return m, suite, continuationManifest(t, "assembly-policy/checkpoint.gooo.fixture", "Checkpoint"),
		continuationManifest(t, "assembly-explainer/main.gooo.fixture", "Explain")
}

func TestResumeWorkspaceNestedHelpersKeepsPolicyRounds(t *testing.T) {
	m, suite, checkpoint, policy := continuationFixture(t)
	ctx := context.Background()
	prior, err := ExecuteWorkspaceWithOptions(ctx, m, suite, ExecuteOptions{AssemblyPolicy: &checkpoint})
	if err != nil || prior.Runtime.FinitePassed != 0 {
		t.Fatal("checkpoint", err)
	}
	prior = savedWorkspaceResult(t, prior)
	next, err := ResumeWorkspace(ctx, m, prior, suite, &policy, "")
	if err != nil || next.Runtime.FinitePassed != 2 || next.Continuation == nil || len(next.PolicyHistory) != 1 {
		t.Fatal("continuation", err)
	}
	counts := next.Composition.Continuation.Activities
	if len(counts) != 2 || counts[1].RecheckedAttempts != 1 || counts[1].AddedAttempts != 1 ||
		next.Continuation.NewModelCalls != 0 || next.Continuation.SavedPolicyStages != 1 {
		t.Fatal("dependency rechecks or policy stage missing", counts, next.Continuation)
	}
	// Mutating the caller's old policy cannot mutate the new snapshot.
	prior.AssemblyPolicy.Manifest.Packages[0].Sources[0].Content += "\n"
	again, err := ResumeWorkspace(ctx, m, savedWorkspaceResult(t, next), suite, &policy, "")
	if err != nil || len(again.PolicyHistory) != 2 || again.Runtime.FinitePassed != 2 {
		t.Fatal("second round", err)
	}
	for _, c := range again.Composition.Continuation.Activities {
		if c.AddedAttempts != 0 || c.RecheckedAttempts != 0 {
			t.Fatal("completed choices consumed new budget", c)
		}
	}
	inputOnly := suite
	inputOnly.Schema = bodyexecution.CompositionInputsSchema
	inputOnly.Cases = []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{"example/tool:Main": json.RawMessage(`17`)}}}
	replay, err := ReplayWorkspace(ctx, m, savedWorkspaceResult(t, again), inputOnly, "")
	if err != nil || replay.Runtime.FiniteTotal != 0 || replay.Runtime.ModelCalls != 0 ||
		replay.Continuation == nil || len(replay.PolicyHistory) != 2 || replay.Replay.ModelCalls != 0 {
		t.Fatal("fresh input replay lost history or invented correctness", err)
	}
}

func TestResumeWorkspaceFromDefaultPolicyAndCancellation(t *testing.T) {
	m, suite, _, policy := continuationFixture(t)
	ctx := context.Background()
	prior, err := ExecuteWorkspace(ctx, m, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	prior = savedWorkspaceResult(t, prior)
	next, err := ResumeWorkspace(ctx, m, prior, suite, &policy, "")
	if err != nil || len(next.PolicyHistory) != 1 || next.PolicyHistory[0] != nil || next.Runtime.FinitePassed != 2 {
		t.Fatal("default original policy did not survive continuation", err)
	}
	if _, err := ReplayWorkspace(ctx, m, savedWorkspaceResult(t, next), suite, ""); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := ResumeWorkspace(canceled, m, prior, suite, &policy, ""); err != context.Canceled {
		t.Fatal("lost cancellation", err)
	}
	if _, err := ResumeWorkspace(nil, m, prior, suite, &policy, ""); err == nil {
		t.Fatal("nil context accepted")
	}
	if _, err := ResumeWorkspace(ctx, m, prior, suite, nil, ""); err == nil {
		t.Fatal("implicit continuation policy accepted")
	}
}

func TestResumeWorkspaceRetainsEarlierFillAndInputObservations(t *testing.T) {
	m, suite := workspaceInputsFixture(t)
	m.Packages[0].Sources[0].Content += `
entity Box id "construction://box" fields { field value id "construction://value" type integer required one }
activity Inspect(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "input + 1" intent "Add one."
 value_case "[1]" -> "{\"value\":2}"
 value_case "[8]" -> "{\"value\":9}"
 attempts "2"
}
bind Lift.result -> Inspect.input
`
	m.Entry.Activity = "Inspect"
	for i := range suite.Cases {
		var value int64
		if err := json.Unmarshal(suite.Cases[i].Expected["example/construction:Lift"], &value); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(map[string]int64{"value": value + 1})
		suite.Cases[i].Expected["example/construction:Inspect"] = raw
	}
	checkpoint := continuationManifest(t, "assembly-policy/checkpoint.gooo.fixture", "Checkpoint")
	policy := continuationManifest(t, "assembly-explainer/main.gooo.fixture", "Explain")
	ctx := context.Background()
	prior, err := ExecuteWorkspaceWithOptions(ctx, m, suite, ExecuteOptions{AssemblyPolicy: &checkpoint})
	if err != nil {
		t.Fatal(err)
	}
	prior = savedWorkspaceResult(t, prior)
	next, err := ResumeWorkspace(ctx, m, prior, suite, &policy, "")
	if err != nil || next.Runtime.FinitePassed != next.Runtime.FiniteTotal || len(next.BodyFills) != 1 ||
		next.Continuation.BodyFillsReplayed != 1 || len(next.Runtime.InputSeparation.EarlierStages) != 1 {
		t.Fatal("earlier fill or input history was lost", err)
	}
	if !reflect.DeepEqual(prior.BodyFills, next.BodyFills) {
		t.Fatal("continuation changed a prior source fill")
	}
	if _, err := ReplayWorkspace(ctx, m, savedWorkspaceResult(t, next), suite, ""); err != nil {
		t.Fatal("continued earlier fill did not replay", err)
	}
}
