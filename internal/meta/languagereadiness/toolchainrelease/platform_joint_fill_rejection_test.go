package toolchainrelease

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func jointFillRejectionFixture(t *testing.T, mode string) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "caller-fill-rejection-"+mode+".json.gz"))
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
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "examples/caller-fill-rejection", name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return raw, read("budget.gooo.fixture"), read("construction-cases.json"), read("evaluation-cases.json")
}

func TestJointFillRejectionPartialCompleteAndReplay(t *testing.T) {
	raw, source, feedback, evaluation := jointFillRejectionFixture(t, "construct")
	selected, err := validateJointFillRejectionSmoke(raw, source, feedback, evaluation, 5, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _, _, _ := jointFillRejectionFixture(t, "replay")
	if got, err := validateJointFillRejectionSmoke(replay, source, feedback, evaluation, 5, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
	if _, err := validateJointFillRejectionSmoke(replay, source, feedback, evaluation, 5, true, ""); err == nil {
		t.Fatal("unbound replay accepted")
	}
	partial, _, _, _ := jointFillRejectionFixture(t, "partial")
	if got, err := validateJointFillRejectionSmoke(partial, source, feedback, evaluation, 3, false, ""); err != nil || got == selected {
		t.Fatal(got, err)
	}
	if _, err := validateJointFillRejectionSmoke(partial, source, feedback, evaluation, 5, false, ""); err == nil {
		t.Fatal("partial result accepted as complete")
	}
	if _, err := validateJointFillRejectionSmoke(raw, append(source, '\n'), feedback, evaluation, 5, false, ""); err == nil {
		t.Fatal("different source accepted")
	}
}

// Decode with UseNumber so the mutation harness cannot round the original integers.
func TestJointFillRejectionRejectsChangedEvidence(t *testing.T) {
	raw, source, feedback, evaluation := jointFillRejectionFixture(t, "construct")
	type edit struct {
		path  []any
		value any
	}
	changes := map[string]edit{
		"schema":                   {[]any{"construction", "schema"}, "gooo/joint-construction/v4"},
		"space":                    {[]any{"construction", "candidate_space"}, "3"},
		"budget":                   {[]any{"construction", "program_budget"}, 3},
		"selection":                {[]any{"construction", "selected_attempt"}, 1},
		"initial rejected omitted": {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "rejected_candidates"}, []any{}},
		"initial scored omitted":   {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "candidate_scores"}, []any{}},
		"initial scoring changed":  {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "candidate_scores", 0, "test_cases_total"}, 2},
		"initial plan":             {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "ir_plan_sha256"}, "changed"},
		"rejection omitted":        {[]any{"construction", "attempts", 1, "rejection"}, nil},
		"rejection stage":          {[]any{"construction", "attempts", 1, "rejection", "stage"}, "NATIVE"},
		"rejection reason":         {[]any{"construction", "attempts", 1, "rejection", "reason"}, "changed"},
		"rejection slot":           {[]any{"construction", "attempts", 1, "rejection", "slot"}, 1},
		"fill rejection omitted":   {[]any{"construction", "attempts", 1, "fill_candidates", 0, "rejection"}, nil},
		"fill rejection stage":     {[]any{"construction", "attempts", 2, "fill_candidates", 0, "rejection", "stage"}, "TYPECHECK"},
		"hole expression":          {[]any{"construction", "attempts", 1, "fill_candidates", 0, "hole_fills", 1, "expression"}, "0"},
		"rejected hole expression": {[]any{"construction", "attempts", 1, "fill_candidates", 0, "rejection", "hole_fills", 1, "expression"}, "0"},
		"candidate id":             {[]any{"construction", "attempts", 1, "fill_candidates", 0, "candidate_id"}, "bounded"},
		"candidate count":          {[]any{"construction", "attempts", 1, "fill_candidates", 0, "candidate_count"}, 3},
		"plan identity":            {[]any{"construction", "attempts", 1, "fill_candidates", 0, "plan_sha256"}, "changed"},
		"source identity":          {[]any{"construction", "attempts", 1, "fill_candidates", 0, "input_source_sha256"}, "changed"},
		"invented fill score":      {[]any{"construction", "attempts", 1, "fill_candidates", 0, "test_cases_total"}, 1},
		"invented local score":     {[]any{"construction", "attempts", 1, "local_passed"}, 1},
		"invented holdout":         {[]any{"construction", "attempts", 1, "fill_candidates", 0, "holdout_cases_total"}, 1},
		"invented execution":       {[]any{"construction", "attempts", 1, "runtime", "stage"}, "COMPLETE"},
		"invented native build":    {[]any{"construction", "attempts", 1, "runtime", "build", "started"}, true},
		"invented process":         {[]any{"construction", "attempts", 1, "runtime", "runs"}, []any{map[string]any{"started": true}}},
		"selector":                 {[]any{"construction", "attempts", 1, "masks", 0}, 2},
		"valid holdout score":      {[]any{"construction", "attempts", 3, "fill_candidates", 0, "holdout_cases_passed"}, 1},
		"caller actual":            {[]any{"construction", "attempts", 3, "runtime", "traces", 0, "deliveries", 0, "actual"}, -8},
		"large integer":            {[]any{"evaluation", "runtime", "traces", 3, "deliveries", 0, "actual"}, json.Number("-9007199254740992")},
		"inference":                {[]any{"evaluation", "new_model_calls"}, 1},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			cursor := value
			for _, key := range change.path[:len(change.path)-1] {
				switch key := key.(type) {
				case string:
					cursor = cursor.(map[string]any)[key]
				case int:
					cursor = cursor.([]any)[key]
				}
			}
			switch key := change.path[len(change.path)-1].(type) {
			case string:
				cursor.(map[string]any)[key] = change.value
			case int:
				cursor.([]any)[key] = change.value
			}
			bad, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := validateJointFillRejectionSmoke(bad, source, feedback, evaluation, 5, false, ""); err == nil {
				t.Fatal("changed release observation accepted")
			}
		})
	}
}
