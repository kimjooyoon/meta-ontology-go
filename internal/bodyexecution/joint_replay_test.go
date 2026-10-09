package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJointReplayRecomputesSelectionsAndCallerEvidence(t *testing.T) {
	ctx := context.Background()
	source := []byte(jointAmbiguousSource)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(ctx, "joint.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	changes := map[string]func(*JointConstruction){
		"mask":        func(r *JointConstruction) { r.Attempts[0].Masks[0] = 1 },
		"local score": func(r *JointConstruction) { r.Attempts[0].LocalPassed = 0 },
		"local value": func(r *JointConstruction) {
			r.Attempts[0].Candidates[0].Cases[0].Actual = json.RawMessage(`{"value":1}`)
		},
		"caller score": func(r *JointConstruction) { r.Attempts[0].Runtime.FinitePassed = 1 },
		"caller value": func(r *JointConstruction) {
			r.Attempts[0].Runtime.Traces[0].Deliveries[0].Actual = json.RawMessage(`6`)
		},
		"selection":       func(r *JointConstruction) { r.SelectedAttempt = 0 },
		"selected source": func(r *JointConstruction) { r.SelectedSource += "\n" },
		"stop":            func(r *JointConstruction) { r.StopReason = "PROGRAM_BUDGET_EXHAUSTED" },
		"caller suite":    func(r *JointConstruction) { r.ConstructionCases.Cases[0].Expected["Main"] = json.RawMessage(`3`) },
		"truncated":       func(r *JointConstruction) { r.Attempts = r.Attempts[:1]; r.SelectedAttempt = 0 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			r, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&r)
			before := compositionDigest(r)
			got, err := ReplayJointComposition(ctx, "joint.gooo", source, r, cases, nativeTool())
			if err == nil {
				t.Fatal("changed observation replayed")
			}
			if name == "caller score" || name == "caller value" {
				f := got.ReplayFailure
				if f == nil || f.AttemptIndex != 0 || f.Stage != "CALLER_COMPARISON" ||
					!sameJointRuntime(f.Runtime, prior.Attempts[0].Runtime) || got.ConstructionReplayed ||
					got.Runtime.Stage != "" || compositionDigest(r) != before {
					t.Fatal("current comparison evidence or original history changed", got)
				}
			}
		})
	}
}

func TestJointSourceAttemptLimitRestrictsEligibleChoices(t *testing.T) {
	source := []byte(strings.Replace(jointAmbiguousSource, `attempts "2"`, `attempts "1"`, 1))
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	r, err := ConstructJointComposition(context.Background(), "limited.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 64, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Attempts) != 1 || r.CandidateSpace != "1" || r.Decision != "PARTIAL_FINITE" || r.StopReason != "DECLARED_SPACE_EXHAUSTED" {
		t.Fatal("whole-program budget expanded the source choice limit", r)
	}
}

func TestJointIndependentCallsAndCancellation(t *testing.T) {
	for i := range 2 {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			t.Parallel()
			source := []byte(jointAmbiguousSource)
			cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
			r, err := ConstructJointComposition(context.Background(), "parallel.gooo", source, cases,
				JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
			if err != nil || r.Decision != "COMPLETE_FINITE" {
				t.Fatal(r.Decision, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := ReplayJointComposition(ctx, "parallel.gooo", source, r, cases, nativeTool()); err == nil {
				t.Fatal("cancelled replay proceeded")
			}
		})
	}
}
