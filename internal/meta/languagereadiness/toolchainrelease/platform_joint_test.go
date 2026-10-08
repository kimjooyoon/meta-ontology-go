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

func jointSmokeFixture(t *testing.T, name string) ([]byte, []byte, []byte) {
	t.Helper()
	root := filepath.Join("..", "..", "..", "..")
	f, err := os.Open(filepath.Join(root, "docs/research/caller-guided-construction-20261008/observations.tar.gz"))
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
			t.Fatal("retained caller observation missing", err)
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
		b, err := os.ReadFile(filepath.Join(root, jointExampleRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return raw, read("construction-cases.json"), read("evaluation-cases.json")
}

func TestJointSmokeRecountsActualCallerConstructionAndReplay(t *testing.T) {
	raw, feedback, evaluation := jointSmokeFixture(t, "simple.json")
	selected, err := validateJointSmoke(raw, feedback, evaluation, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _, _ := jointSmokeFixture(t, "simple-replay.json")
	if got, err := validateJointSmoke(replay, feedback, evaluation, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
	for _, selected := range []string{"", "sha256:other"} {
		if _, err := validateJointSmoke(replay, feedback, evaluation, true, selected); err == nil {
			t.Fatal("unbound caller replay accepted")
		}
	}
}

func TestJointSmokeRequiresOriginalLocalAndIndependentCallerValues(t *testing.T) {
	raw, feedback, evaluation := jointSmokeFixture(t, "simple.json")
	changes := map[string]func(*jointSmokeOutput){
		"missing generated flag": func(r *jointSmokeOutput) { r.Generated = nil },
		"partial":                func(r *jointSmokeOutput) { r.Construction.Decision = "PARTIAL_FINITE" },
		"wrong stop":             func(r *jointSmokeOutput) { r.Construction.StopReason = "PROGRAM_BUDGET_EXHAUSTED" },
		"failed":                 func(r *jointSmokeOutput) { r.Construction.Failure = "failed" },
		"missing budget":         func(r *jointSmokeOutput) { r.Construction.Budget = nil },
		"wrong selection":        func(r *jointSmokeOutput) { *r.Construction.SelectedAttempt = 0 },
		"missing model status":   func(r *jointSmokeOutput) { r.Construction.Initial.Model.Loaded = nil },
		"model loaded":           func(r *jointSmokeOutput) { *r.Construction.Initial.Model.Loaded = true },
		"missing attempt":        func(r *jointSmokeOutput) { r.Construction.Attempts = r.Construction.Attempts[:1] },
		"wrong mask":             func(r *jointSmokeOutput) { r.Construction.Attempts[0].Masks[0] = 1 },
		"local missing":          func(r *jointSmokeOutput) { r.Construction.Attempts[0].Candidates = nil },
		"local count":            func(r *jointSmokeOutput) { *r.Construction.Attempts[0].Passed = 0 },
		"local value": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].Candidates[0].Cases[0].Actual = json.RawMessage(`{"value":1}`)
		},
		"baseline false success": func(r *jointSmokeOutput) { *r.Construction.Attempts[0].Runtime.Passed = 1 },
		"baseline changed": func(r *jointSmokeOutput) {
			r.Construction.Attempts[0].Runtime.Traces[0].Deliveries[0].Actual = json.RawMessage(`6`)
		},
		"new prediction":     func(r *jointSmokeOutput) { *r.Evaluation.Calls = 1 },
		"missing call count": func(r *jointSmokeOutput) { r.Evaluation.Runtime.Calls = nil },
		"runtime failure":    func(r *jointSmokeOutput) { r.Evaluation.Runtime.Failure = "failed" },
		"runtime replay":     func(r *jointSmokeOutput) { *r.Evaluation.Runtime.Replay = false },
		"history replay":     func(r *jointSmokeOutput) { *r.Evaluation.Replayed = true },
		"consumed input":     func(r *jointSmokeOutput) { *r.Evaluation.Separation.Consumed = 0 },
		"different program":  func(r *jointSmokeOutput) { r.Evaluation.Runtime.SHA = "sha256:other" },
		"missing row":        func(r *jointSmokeOutput) { r.Evaluation.Runtime.Traces = r.Evaluation.Runtime.Traces[:3] },
		"duplicate row":      func(r *jointSmokeOutput) { r.Evaluation.Runtime.Traces[0].Index = 1 },
		"changed input":      func(r *jointSmokeOutput) { r.Evaluation.Runtime.Traces[0].Deliveries[0].Input = json.RawMessage(`4`) },
		"large integer": func(r *jointSmokeOutput) {
			r.Evaluation.Runtime.Traces[3].Deliveries[0].Actual = json.RawMessage(`18014398509481985`)
		},
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
			if _, err := validateJointSmoke(bad, feedback, evaluation, false, ""); err == nil {
				t.Fatal("changed caller observation accepted")
			}
		})
	}
	for _, reference := range [][]byte{nil, []byte(`{"schema":"other","cases":[]}`), feedback} {
		if _, err := validateJointSmoke(raw, feedback, reference, false, ""); err == nil {
			t.Fatal("wrong independent evaluation accepted")
		}
	}
}
