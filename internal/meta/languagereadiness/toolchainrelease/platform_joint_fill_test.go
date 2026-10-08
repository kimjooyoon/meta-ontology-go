package toolchainrelease

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func jointFillSmokeFixture(t *testing.T, mode string) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "caller-fill-"+mode+".json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", jointFillExampleRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return raw, read("budget.gooo.fixture"), read("construction-cases.json"), read("evaluation-cases.json")
}

func TestJointFillSmokeRetainsPartialCompleteAndReplay(t *testing.T) {
	raw, source, feedback, evaluation := jointFillSmokeFixture(t, "construct")
	selected, err := validateJointFillSmoke(raw, source, feedback, evaluation, 3, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _, _, _ := jointFillSmokeFixture(t, "replay")
	if got, err := validateJointFillSmoke(replay, source, feedback, evaluation, 3, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
	if _, err := validateJointFillSmoke(replay, source, feedback, evaluation, 3, true, ""); err == nil {
		t.Fatal("unbound replay accepted")
	}
	partial, _, _, _ := jointFillSmokeFixture(t, "partial")
	if got, err := validateJointFillSmoke(partial, source, feedback, evaluation, 2, false, ""); err != nil || got == selected {
		t.Fatal(got, err)
	}
	if _, err := validateJointFillSmoke(partial, source, feedback, evaluation, 3, false, ""); err == nil {
		t.Fatal("partial result accepted as complete")
	}
	if _, err := validateJointFillSmoke(raw, append(source, '\n'), feedback, evaluation, 3, false, ""); err == nil {
		t.Fatal("different source accepted")
	}
}

func TestJointFillSmokeRejectsChangedAssignmentsAndCounts(t *testing.T) {
	raw, source, feedback, evaluation := jointFillSmokeFixture(t, "construct")
	changes := map[string]func(*jointSmokeOutput){
		"schema":            func(r *jointSmokeOutput) { r.Construction.Schema = "gooo/joint-construction/v3" },
		"kind":              func(r *jointSmokeOutput) { r.Construction.Kinds[0] = "record_mask" },
		"selector":          func(r *jointSmokeOutput) { r.Construction.Attempts[0].Masks[0] = 1 },
		"omitted candidate": func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates = nil },
		"initial winner": func(r *jointSmokeOutput) {
			r.Construction.Initial.Preparations[0].Generation.Report.Fill.Selected = "bounded"
		},
		"false initial prediction": func(r *jointSmokeOutput) { *r.Construction.Initial.Preparations[0].Generation.Report.Fill.Calls = 1 },
		"candidate id":             func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates[0].ID = "bounded" },
		"hole id":                  func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates[0].Holes[0].ID = "cap" },
		"hole expression":          func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates[0].Holes[0].Expression = "true" },
		"plan identity":            func(r *jointSmokeOutput) { r.Construction.Attempts[1].FillCandidates[0].PlanSHA = "changed" },
		"source identity":          func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates[0].InputSHA = "changed" },
		"selected source":          func(r *jointSmokeOutput) { r.Construction.Attempts[2].FillCandidates[0].SelectedSHA = "changed" },
		"false holdout success":    func(r *jointSmokeOutput) { *r.Construction.Attempts[0].FillCandidates[0].HoldoutPassed = 1 },
		"missing holdout count":    func(r *jointSmokeOutput) { r.Construction.Attempts[0].FillCandidates[0].HoldoutTotal = nil },
		"holdout in training":      func(r *jointSmokeOutput) { *r.Construction.Attempts[0].Total = 2 },
		"holdout expected": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].FillCandidates[0].ValueHoldout[0].Expected = []byte(`{"next":9,"exhausted":true}`)
		},
		"local actual": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].FillCandidates[0].Values[0].Actual = []byte(`{"next":0,"exhausted":false}`)
		},
		"caller actual": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].Runtime.Traces[0].Deliveries[0].Actual = []byte("-8")
		},
		"large actual": func(r *jointSmokeOutput) {
			r.Evaluation.Runtime.Traces[3].Deliveries[0].Actual = []byte("-9007199254740992")
		},
		"large input": func(r *jointSmokeOutput) {
			r.Evaluation.Runtime.Traces[3].Deliveries[0].Input = []byte(`{"used":9007199254740992,"limit":9007199254740993}`)
		},
		"model calls":      func(r *jointSmokeOutput) { *r.Evaluation.Calls = 1 },
		"selected program": func(r *jointSmokeOutput) { r.Evaluation.Runtime.SHA = "changed" },
		"selection":        func(r *jointSmokeOutput) { *r.Construction.SelectedAttempt = 0 },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			var r jointSmokeOutput
			if err := json.Unmarshal(raw, &r); err != nil {
				t.Fatal(err)
			}
			change(&r)
			bad, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := validateJointFillSmoke(bad, source, feedback, evaluation, 3, false, ""); err == nil {
				t.Fatal("changed release observation accepted")
			}
		})
	}
}
