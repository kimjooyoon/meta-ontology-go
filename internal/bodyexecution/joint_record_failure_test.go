package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJointRecordEvaluationFailureContinuesAndReplays(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile("../../examples/caller-record-failures/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	decode := func(name string) CompositionCases {
		t.Helper()
		suite, err := DecodeCompositionCases(read(name))
		if err != nil {
			t.Fatal(err)
		}
		return suite
	}
	source := read("main.gooo.fixture")
	evaluation := decode("evaluation-cases.json")
	for _, callerMatches := range []bool{true, false} {
		construction := decode("construction-cases.json")
		if !callerMatches {
			construction.Cases[0].Expected["Main"] = json.RawMessage("100")
		}
		ctx := context.Background()
		prior, got, err := ConstructAndEvaluateJoint(ctx, "record.gooo", source, construction, evaluation,
			JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
		if err != nil || got.Runtime.FinitePassed != 3 || got.NewModelCalls != 0 {
			t.Fatal("record recovery did not reach native execution", got, err)
		}
		if callerMatches {
			if prior.Decision != "COMPLETE_FINITE" || len(prior.Attempts) != 1 {
				t.Fatal("valid local alternative was not selected", prior.Decision)
			}
		} else if prior.Decision != "PARTIAL_FINITE" || len(prior.Attempts) != 2 ||
			prior.Attempts[1].Rejection == nil || prior.Attempts[1].Rejection.Stage != "LOCAL_RECORD_ASSEMBLY" ||
			prior.Attempts[1].Runtime.Stage != "" || prior.Schema != jointRejectionSchema {
			t.Fatal("failed explicit record candidate did not consume one bounded attempt", prior)
		}
		raw, _ := json.MarshalIndent(prior, "", "  ")
		saved, err := DecodeJointConstruction(raw)
		if err != nil {
			t.Fatal(err)
		}
		replayed, err := ReplayJointComposition(ctx, "record.gooo", source, saved, evaluation, nativeTool())
		if err != nil || !replayed.ConstructionReplayed || replayed.Runtime.FinitePassed != 3 || replayed.NewModelCalls != 0 {
			t.Fatal("native record rejection history did not replay", replayed, err)
		}
	}
}

func TestJointRecordFailureRetainsInitialAttemptPrefix(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-record-failures/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), `attempts "2"`, `attempts "1"`, 1))
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":4}}]`)
	r, err := ConstructJointComposition(context.Background(), "record.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	failure := r.Initial.RecordFailure
	if err == nil || r.Initial.Stage != "ACTIVITY_GENERATION" || failure == nil ||
		len(failure.Attempts) != 1 || len(r.Attempts) != 0 || failure.Total != 0 {
		t.Fatal("initial record failure lost its bounded observation", r, err)
	}
	failed := failure.Attempts[0]
	if failed.Status != "EVALUATION_FAILED" || failed.Total != 0 || failed.LocalCalls == nil ||
		failed.LocalCalls.EvaluationCompleted || len(failed.LocalCalls.Calls) != 2 ||
		string(failed.LocalCalls.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed initial composition lost actual helper calls", failed)
	}
	if err := VerifyComposition(context.Background(), "record.gooo", source, r.Initial); err == nil {
		t.Fatal("failed record composition became a completed construction")
	}
}
