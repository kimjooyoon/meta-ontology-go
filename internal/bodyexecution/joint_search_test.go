package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func jointSearchSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/caller-ir-search/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestJointCallerSelectsSourceIRExpression(t *testing.T) {
	source := jointSearchSource(t)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":3}}]`)
	prior, err := ConstructJointComposition(context.Background(), "search.gooo", source, cases, JointOptions{EntryActivity: "Main", ProgramBudget: 8, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if prior.Schema != "gooo/joint-construction/v2" || prior.Decision != "COMPLETE_FINITE" || len(prior.Attempts) != 2 || prior.CandidateSpace != "5" {
		t.Fatal(prior)
	}
	if prior.Initial.Preparations[0].Generation.Report.BodySearch.SelectedExpression != "input" || strings.Contains(prior.SelectedSource, "__GOOO_BODY_HOLE_") || strings.Contains(prior.SelectedSource, "assembling") {
		t.Fatal("caller did not reopen the completed local expression")
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":7},"expected":{"Main":7}},{"inputs":{"Main":-5},"expected":{"Main":-5}},{"inputs":{"Main":9007199254740993},"expected":{"Main":9007199254740993}}]`)
	raw, _ := json.MarshalIndent(prior, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(context.Background(), "search.gooo", source, saved, evaluation, nativeTool())
	if err != nil || replay.Runtime.FinitePassed != 3 || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.InputSeparation.OtherInputs != 3 {
		t.Fatal(replay, err)
	}
}

func TestJointSearchLimitsConflictAndBinding(t *testing.T) {
	original := string(jointSearchSource(t))
	for _, tc := range []struct {
		name, source, cases, space, decision string
		budget, attempts                     int
	}{
		{"program limit", original, `[{"inputs":{"Main":3},"expected":{"Main":3}}]`, "5", "PARTIAL_FINITE", 1, 1},
		{"source limit", strings.Replace(original, `attempts "5"`, `attempts "1"`, 1), `[{"inputs":{"Main":3},"expected":{"Main":3}}]`, "1", "PARTIAL_FINITE", 8, 1},
		{"local conflict", strings.Replace(original, `case "0" -> "0"`, `case "1" -> "0"`, 1), `[{"inputs":{"Main":3},"expected":{"Main":6}}]`, "5", "PARTIAL_FINITE", 8, 5},
		{"explicit binding", strings.Replace(original, `activity Main(Integer) -> Integer computes "return Choose(input) + input"`, `activity Main(Integer) -> Integer computes "return input"`+"\nbind Choose.result -> Main.input", 1), `[{"inputs":{"Choose":3},"expected":{"Main":0}}]`, "5", "COMPLETE_FINITE", 8, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := "Main"
			if tc.name == "explicit binding" {
				entry = ""
			}
			cases := jointCases(t, tc.cases)
			prior, err := ConstructJointComposition(context.Background(), "bounded.gooo", []byte(tc.source), cases, JointOptions{EntryActivity: entry, ProgramBudget: tc.budget, GoBinary: nativeTool()})
			if err != nil {
				t.Fatal(err)
			}
			if len(prior.Attempts) != tc.attempts || prior.CandidateSpace != tc.space || prior.Decision != tc.decision {
				t.Fatal(prior)
			}
			if tc.name == "local conflict" && (prior.SelectedAttempt != 0 || prior.Attempts[1].LocalPassed != 0 || prior.Attempts[1].Runtime.FinitePassed != 1) {
				t.Fatal("local obligation was erased")
			}
			if _, err = ReplayJointComposition(context.Background(), "bounded.gooo", []byte(tc.source), prior, cases, nativeTool()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestJointSearchAndRecordChoicesCompose(t *testing.T) {
	source := strings.Replace(string(jointSearchSource(t)), `activity Main(Integer) -> Integer computes "return Choose(input) + input"`, `entity Box id "callerhole://box" fields { field value id "callerhole://value" type integer required one }
activity Pick(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "double" field_value at "0" alternative "input * 2" intent "Double the value."
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let n = Choose(input); let p = Pick(input); if n == 0 { return p.value }; return -1"`, 1)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(context.Background(), "mixed.gooo", []byte(source), cases, JointOptions{EntryActivity: "Main", ProgramBudget: 10, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if prior.CandidateSpace != "10" || prior.Decision != "COMPLETE_FINITE" || len(prior.CandidateKinds) != 2 {
		t.Fatal(prior)
	}
	for _, attempt := range prior.Attempts {
		if len(attempt.Candidates) != 1 || len(attempt.SearchCandidates) != 1 || attempt.LocalPassed != 2 || attempt.LocalTotal != 2 {
			t.Fatal(attempt)
		}
	}
	if _, err = ReplayJointComposition(context.Background(), "mixed.gooo", []byte(source), prior, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}

func TestJointSearchReplayRechecksTypedSelectionAndCases(t *testing.T) {
	source := jointSearchSource(t)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":3}}]`)
	prior, err := ConstructJointComposition(context.Background(), "search.gooo", source, cases, JointOptions{EntryActivity: "Main", ProgramBudget: 8, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	for name, change := range map[string]func(*JointConstruction){
		"kind":                func(r *JointConstruction) { r.CandidateKinds[0] = "record_mask" },
		"schema":              func(r *JointConstruction) { r.Schema = jointSchema },
		"expression":          func(r *JointConstruction) { r.Attempts[0].SearchCandidates[0].Attempt.Expression = "999" },
		"actual":              func(r *JointConstruction) { r.Attempts[0].SearchCandidates[0].Attempt.CaseResults[0].Actual = 1 },
		"omitted grammar":     func(r *JointConstruction) { r.Attempts[0].SearchCandidates[0].Generation.CandidatesOmitted++ },
		"missing observation": func(r *JointConstruction) { r.Attempts[0].SearchCandidates = nil },
		"index":               func(r *JointConstruction) { r.Attempts[0].Masks[0] = 4 },
	} {
		t.Run(name, func(t *testing.T) {
			saved, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&saved)
			if _, err = ReplayJointComposition(context.Background(), "search.gooo", source, saved, cases, nativeTool()); err == nil {
				t.Fatal("altered search observation replayed")
			}
		})
	}
}
