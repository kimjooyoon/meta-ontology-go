package toolchainrelease

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func jointSearchSmokeFixture(t *testing.T, name string) ([]byte, []byte, []byte) {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..")
	f, err := os.Open(filepath.Join(root, "docs/research/caller-search-rejection-20261009/observations.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	r := tar.NewReader(g)
	var raw []byte
	for {
		header, err := r.Next()
		if err != nil {
			t.Fatal("retained search observation missing", err)
		}
		if header.Name == name {
			raw, err = io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(root, jointSearchExampleRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return raw, read("construction-cases.json"), read("evaluation-cases.json")
}

func TestJointSearchSmokeRecountsRejectionAndNativeValues(t *testing.T) {
	raw, feedback, evaluation := jointSearchSmokeFixture(t, "scalar.json")
	selected, err := validateJointSearchSmoke(raw, feedback, evaluation, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _, _ := jointSearchSmokeFixture(t, "scalar-replay.json")
	if got, err := validateJointSearchSmoke(replay, feedback, evaluation, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
	if _, err := validateJointSearchSmoke(replay, feedback, evaluation, true, ""); err == nil {
		t.Fatal("unbound replay accepted")
	}
	partial, _, _ := jointSearchSmokeFixture(t, "partial.json")
	if _, err := validateJointSearchSmoke(partial, feedback, evaluation, false, ""); err == nil {
		t.Fatal("partial candidate passed release regression")
	}
}

func TestJointSearchSmokeRejectsChangedHistory(t *testing.T) {
	raw, feedback, evaluation := jointSearchSmokeFixture(t, "scalar.json")
	changes := map[string]func(*jointSmokeOutput){
		"version":            func(r *jointSmokeOutput) { r.Construction.Schema = "gooo/joint-construction/v2" },
		"kind":               func(r *jointSmokeOutput) { r.Construction.Kinds[0] = "record_mask" },
		"selection":          func(r *jointSmokeOutput) { *r.Construction.SelectedAttempt = 1 },
		"missing rejection":  func(r *jointSmokeOutput) { r.Construction.Attempts[1].Rejection = nil },
		"missing slot":       func(r *jointSmokeOutput) { r.Construction.Attempts[1].Rejection.Slot = nil },
		"reason":             func(r *jointSmokeOutput) { r.Construction.Attempts[1].Rejection.Reason = "changed" },
		"false score":        func(r *jointSmokeOutput) { *r.Construction.Attempts[1].SearchCandidates[0].Attempt.Scored = true },
		"missing score flag": func(r *jointSmokeOutput) { r.Construction.Attempts[1].SearchCandidates[0].Attempt.Scored = nil },
		"false execution":    func(r *jointSmokeOutput) { r.Construction.Attempts[1].Runtime.Stage = "COMPLETE" },
		"local actual": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].SearchCandidates[0].Attempt.Cases[0].Actual = []byte("1")
		},
		"first actual": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].Runtime.Traces[0].Deliveries[0].Actual = []byte("-1")
		},
		"wrong large input": func(r *jointSmokeOutput) {
			r.Evaluation.Runtime.Traces[3].Deliveries[0].Input = []byte("9007199254740992")
		},
		"wrong actual":    func(r *jointSmokeOutput) { r.Evaluation.Runtime.Traces[3].Deliveries[0].Actual = []byte("1") },
		"new prediction":  func(r *jointSmokeOutput) { *r.Evaluation.Calls = 1 },
		"input exposure":  func(r *jointSmokeOutput) { *r.Evaluation.Separation.Consumed = 1 },
		"program changed": func(r *jointSmokeOutput) { r.Evaluation.Runtime.SHA = "different" },
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
			if _, err := validateJointSearchSmoke(bad, feedback, evaluation, false, ""); err == nil {
				t.Fatal("changed search release observation accepted")
			}
		})
	}
}
