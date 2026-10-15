package toolchainrelease

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func helperFillFixture(t *testing.T, mode string) ([]byte, [][]byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "helper-fill-history-"+mode+".json.gz"))
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
	files := make([][]byte, 3)
	for i, name := range []string{"main.gooo.fixture", "construction-cases.json", "evaluation-cases.json"} {
		files[i], err = os.ReadFile(filepath.Join("../../../..", helperFillRoot, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	return raw, files
}

func TestHelperFillReleaseRetainsFailurePreflightAndReplay(t *testing.T) {
	raw, files := helperFillFixture(t, "construct")
	retained, err := validateHelperFillHistory(raw, files[0], files[1], files[2], false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _ := helperFillFixture(t, "replay")
	if got, err := validateHelperFillHistory(replay, files[0], files[1], files[2], true, retained); err != nil || got != retained {
		t.Fatal("complete construction was not retained", err)
	}
	if _, err := validateHelperFillHistory(replay, files[0], files[1], files[2], true, ""); err == nil {
		t.Fatal("replay without original history passed")
	}
	if _, err := validateHelperFillHistory(raw, append(bytes.Clone(files[0]), '\n'), files[1], files[2], false, ""); err == nil {
		t.Fatal("different source passed")
	}
	changed := bytes.ReplaceAll(files[2], []byte("\"Main\":4"), []byte("\"Main\":5"))
	if bytes.Equal(changed, files[2]) {
		t.Fatal("case mutation missed its target")
	}
	if _, err := validateHelperFillHistory(raw, files[0], files[1], changed, false, ""); err == nil {
		t.Fatal("native results were not compared with the independent case file")
	}
	changed = changedNativeSmoke(t, replay, nativeSmokeEdit{[]any{"construction", "elapsed_ns"}, 0})
	if _, err := validateHelperFillHistory(changed, files[0], files[1], files[2], true, retained); err == nil {
		t.Fatal("replay replaced original construction values")
	}
}

func TestHelperFillReleaseRejectsIncompleteOrChangedEvidence(t *testing.T) {
	raw, files := helperFillFixture(t, "construct")
	prefix := []any{"construction", "initial", "preparations", 1, "generation", "report", "body_fill"}
	field := func(suffix ...any) []any { return append(slices.Clone(prefix), suffix...) }
	for name, edit := range map[string]nativeSmokeEdit{
		"native value":                   {[]any{"evaluation", "runtime", "traces", 0, "deliveries", 0, "actual"}, 99},
		"new inference":                  {[]any{"evaluation", "new_model_calls"}, 1},
		"missing inference count":        {[]any{"evaluation", "new_model_calls"}, nil},
		"lost overlap":                   {[]any{"evaluation", "runtime", "input_separation", "overlapping_inputs"}, 0},
		"unknown erased":                 {[]any{"evaluation", "runtime", "input_separation", "unknown_inputs"}, 0},
		"missing preflight":              {[]any{"construction", "initial", "fill_validations"}, nil},
		"preflight checkpoint":           {[]any{"construction", "initial", "fill_validations", 0, "observation", "source_sha256"}, "changed"},
		"preflight completion":           {[]any{"construction", "initial", "fill_validations", 0, "observation", "completed"}, false},
		"missing rejected prefix":        {field("local_calls", "evaluations", 0, "calls"), nil},
		"changed rejected input":         {field("local_calls", "evaluations", 0, "calls", 1, "inputs"), []any{99}},
		"failed pass completed":          {field("local_calls", "evaluations", 0, "evaluation_completed"), true},
		"failed pass scored":             {field("local_calls", "evaluations", 0, "cases_completed"), 2},
		"rejection erased":               {field("rejected_candidates"), nil},
		"failed candidate score":         {field("candidate_scores", 0, "id"), "fails"},
		"changed holdout":                {field("holdout_case_results", 0, "actual"), 99},
		"missing local prediction count": {field("local_model_predictions"), nil},
		"missing provider count":         {field("external_provider_calls"), nil},
		"provider count unknown":         {field("external_provider_calls_known"), false},
		"caller source digest":           {[]any{"construction", "attempts", 0, "fill_candidates", 0, "selected_source_sha256"}, "changed"},
		"changed helper identity":        {field("local_calls", "evaluations", 0, "calls", 0, "program_sha256"), "changed"},
		"missing caller history":         {[]any{"construction", "attempts", 0, "fill_candidates", 0, "local_calls"}, nil},
	} {
		t.Run(name, func(t *testing.T) {
			changed := changedNativeSmoke(t, raw, edit)
			if _, err := validateHelperFillHistory(changed, files[0], files[1], files[2], false, ""); err == nil {
				t.Fatal("changed release evidence passed")
			}
		})
	}
}

func TestHelperFillReleaseCommandsAndFailureArtifacts(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	construct, _ := helperFillFixture(t, "construct")
	replay, _ := helperFillFixture(t, "replay")
	for _, failed := range []bool{false, true} {
		input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "fixture"}}
		work, calls := t.TempDir(), 0
		stop := errors.New("native process failed")
		run := func(dir string, env []string, name string, args ...string) ([]byte, error) {
			if dir != root || len(env) != 0 || name != "packaged-gooo" {
				t.Fatal("wrong compiler invocation")
			}
			want := []string{"body-construct", "--source", filepath.Join(helperFillRoot, "main.gooo.fixture"),
				"--cases", filepath.Join(helperFillRoot, "evaluation-cases.json"), "--format", "json"}
			out := filepath.Join(work, "helper-fill-history")
			if calls == 0 {
				want = append(want, "--entry", "Main", "--attempts", "4", "--construction-cases",
					filepath.Join(helperFillRoot, "construction-cases.json"), "--out", out)
			} else {
				want = append(want, "--construction", filepath.Join(out, "construction.json"))
			}
			if !slices.Equal(args, want) {
				t.Fatal("wrong source/case/replay arguments", args)
			}
			calls++
			if failed {
				return construct, stop
			}
			if calls == 1 {
				return construct, nil
			}
			return replay, nil
		}
		err := runHelperFillHistory("packaged-gooo", work, input, run)
		if failed && (!errors.Is(err, stop) || calls != 1) || !failed && (err != nil || calls != 2) {
			t.Fatal("native failure or command order changed", calls, err)
		}
		got, err := os.ReadFile(filepath.Join(input.OutputDir, "fixture-helper-fill-history-construct.json"))
		if err != nil || !bytes.Equal(got, construct) {
			t.Fatal("original output was not retained", err)
		}
	}
}
