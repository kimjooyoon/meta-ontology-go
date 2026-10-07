package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func assemblyPolicyFixture(t *testing.T) RecordAssemblyPolicy {
	t.Helper()
	raw, err := os.ReadFile("../../examples/assembly-explainer/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return RecordAssemblyPolicy{Source: string(raw), Activity: "Explain"}
}

func fixedPolicy(t *testing.T, state, operation string) RecordAssemblyPolicy {
	t.Helper()
	policy := assemblyPolicyFixture(t)
	start := strings.Index(policy.Source, "computes `")
	policy.Source = policy.Source[:start] + "computes `return Explanation{state: " +
		quoted(state) + ", next_operation: " + quoted(operation) + ", message: \"fixed experiment\"}`\n"
	return policy
}

func quoted(value string) string { raw, _ := json.Marshal(value); return string(raw) }

func TestRecordPolicyControlsTheNextCandidate(t *testing.T) {
	ctx := context.Background()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	source := recordUpdatesFixture(t)
	for _, test := range []struct {
		name     string
		policy   RecordAssemblyPolicy
		complete bool
	}{
		{"explain", assemblyPolicyFixture(t), true},
		{"stop", fixedPolicy(t, "PASS", "USE_OBSERVED_CANDIDATE"), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := g.GenerateRecordAssemblyWithPolicy(ctx, "record.gooo", source, "Select", test.policy)
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r.Control == nil || len(r.Control.Decisions) != len(r.Attempts) || r.ModelCalls != 0 {
				t.Fatal("policy was not evaluated between candidates", r)
			}
			if test.complete && (r.Passed != r.Total || len(r.Attempts) < 2 || !r.Control.Decisions[0].Continue) {
				t.Fatal("Gooo did not continue from its failed attempt", r)
			}
			if !test.complete && (len(r.Attempts) != 1 || r.Status != "PARTIAL_FINITE" || r.Passed == r.Total) {
				t.Fatal("policy stop or claimed PASS changed the measured result", r)
			}
			replayed, err := RealizeSourceAssembly(ctx, "record.gooo", source, result)
			if err != nil || replayed.ModelCalls != 0 || replayed.Source != result.GoooSource {
				t.Fatal("policy-controlled construction did not replay", err)
			}
			r.Control.Decisions[0].Operation = "OBSERVE_NEW_INPUTS"
			if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, result); err == nil {
				t.Fatal("changed policy decision replayed")
			}
		})
	}
}

func TestRecordPolicyKeepsAnUnsolvedSpacePartial(t *testing.T) {
	g, _ := NewTypedPathGenerator("")
	source := recordUpdatesFixture(t)
	source = []byte(strings.Replace(string(source), `alternative "\"ready\""`, `alternative "\"unresolved\""`, 1))
	result, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "unsolved.gooo", source, "Select", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Status != "PARTIAL_FINITE" || r.Passed == r.Total || len(r.Attempts) != 8 ||
		r.Control.Decisions[7].Operation != "EXPAND_DECLARED_CHOICES" || r.Control.Decisions[7].Continue {
		t.Fatal("unresolved space was hidden", r)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "unsolved.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordPolicyTypedRejectionAndSingleModelCall(t *testing.T) {
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	source := recordUnusedCombinationFixture(t)
	result, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "typed.gooo", source, "Select",
		fixedPolicy(t, "PROGRESS", "CONTINUE_CANDIDATES"))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 1 || r.FieldsPassed != 15 || len(r.Control.Decisions) != len(r.Attempts)-1 {
		t.Fatal("type rejection acquired a score or repeated model inference", r)
	}
	for _, d := range r.Control.Decisions {
		if d.Input.Scored > d.Input.Budget || r.Attempts[d.AttemptIndex].Total == 0 {
			t.Fatal("policy received unscored counts", d)
		}
	}
	if _, err := RealizeSourceAssembly(context.Background(), "typed.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordPolicyRejectsInvalidContractAndOperation(t *testing.T) {
	g, _ := NewTypedPathGenerator("")
	bad := assemblyPolicyFixture(t)
	bad.Source = strings.ReplaceAll(bad.Source, "gooo://tools/assembly-observation", "other://observation")
	for _, policy := range []RecordAssemblyPolicy{bad, fixedPolicy(t, "PASS", "INVENT_SOURCE")} {
		if _, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "record.gooo", recordUpdatesFixture(t), "Select", policy); err == nil {
			t.Fatal("invalid policy was accepted")
		}
	}
}

func TestRecordPolicyRoundTripAndBoundTrace(t *testing.T) {
	g, _ := NewTypedPathGenerator("")
	source := recordUpdatesFixture(t)
	result, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "record.gooo", source, "Select", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*RecordAssemblyReceipt){
		func(r *RecordAssemblyReceipt) {},
		func(r *RecordAssemblyReceipt) { r.Control.SourceSHA256 = "changed" },
		func(r *RecordAssemblyReceipt) { r.Control.Decisions[0].Input.Best++ },
		func(r *RecordAssemblyReceipt) { r.Control.Policy.Source += "\n" },
		func(r *RecordAssemblyReceipt) { r.Control = nil },
	} {
		var copied Result
		if err = json.Unmarshal(raw, &copied); err != nil {
			t.Fatal(err)
		}
		change(copied.Report.RecordAssembly)
		_, replayErr := RealizeSourceAssembly(context.Background(), "record.gooo", source, copied)
		unchanged := copied.Report.RecordAssembly.Control != nil &&
			copied.Report.RecordAssembly.Control.SourceSHA256 == result.Report.RecordAssembly.Control.SourceSHA256 &&
			copied.Report.RecordAssembly.Control.Policy.Source == result.Report.RecordAssembly.Control.Policy.Source &&
			copied.Report.RecordAssembly.Control.Decisions[0].Input == result.Report.RecordAssembly.Control.Decisions[0].Input
		if (replayErr == nil) != unchanged {
			t.Fatal("policy replay binding", replayErr)
		}
	}
}

func TestRecordPolicyRequestsKeepIndependentState(t *testing.T) {
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	source := recordUpdatesFixture(t)
	policy := assemblyPolicyFixture(t)
	errors := make(chan error, 4)
	for range 4 {
		go func() {
			result, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "record.gooo", source, "Select", policy)
			if err == nil {
				r := result.Report.RecordAssembly
				if r.ModelCalls != 1 || r.FieldsPassed != 15 || len(r.Control.Decisions) != len(r.Attempts) {
					err = fmt.Errorf("request borrowed another construction state")
				}
			}
			errors <- err
		}()
	}
	for range 4 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
}
