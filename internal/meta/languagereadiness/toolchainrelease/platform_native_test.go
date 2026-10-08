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

func nativeSmokeFixture(t *testing.T, mode string) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "native-arithmetic-"+mode+".json.gz"))
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
		b, err := os.ReadFile(filepath.Join("../../../../", nativeArithmeticRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	if mode == "graph-construct" || mode == "graph-replay" {
		return raw, read("graph.gooo.fixture"), nil, read("graph-cases.json")
	}
	return raw, read("source.gooo.fixture"), read("construction-cases.json"), read("evaluation-cases.json")
}

func TestNativeArithmeticReleasePartialCompleteReplayAndGraph(t *testing.T) {
	raw, source, feedback, evaluation := nativeSmokeFixture(t, "construct")
	selected, err := validateNativeArithmeticSmoke(raw, source, feedback, evaluation, 6, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay, _, _, _ := nativeSmokeFixture(t, "replay")
	if got, err := validateNativeArithmeticSmoke(replay, source, feedback, evaluation, 6, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
	partial, _, _, _ := nativeSmokeFixture(t, "partial")
	if got, err := validateNativeArithmeticSmoke(partial, source, feedback, evaluation, 5, false, ""); err != nil || got == selected {
		t.Fatal(got, err)
	}
	if _, err := validateNativeArithmeticSmoke(partial, source, feedback, evaluation, 6, false, ""); err == nil {
		t.Fatal("partial accepted as complete")
	}
	if _, err := validateNativeArithmeticSmoke(replay, source, feedback, evaluation, 6, true, ""); err == nil {
		t.Fatal("unbound replay accepted")
	}
	if _, err := validateNativeArithmeticSmoke(raw, append(source, '\n'), feedback, evaluation, 6, false, ""); err == nil {
		t.Fatal("changed source accepted")
	}
	graph, graphSource, _, graphCases := nativeSmokeFixture(t, "graph-construct")
	selected, err = validateNativeArithmeticGraph(graph, graphSource, graphCases, false, "")
	if err != nil {
		t.Fatal(err)
	}
	graphReplay, _, _, _ := nativeSmokeFixture(t, "graph-replay")
	if got, err := validateNativeArithmeticGraph(graphReplay, graphSource, graphCases, true, selected); err != nil || got != selected {
		t.Fatal(got, err)
	}
}

type nativeSmokeEdit struct {
	path  []any
	value any
}

func changedNativeSmoke(t *testing.T, raw []byte, edit nativeSmokeEdit) []byte {
	t.Helper()
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&value); err != nil {
		t.Fatal(err)
	}
	cursor := value
	for _, key := range edit.path[:len(edit.path)-1] {
		switch k := key.(type) {
		case string:
			cursor = cursor.(map[string]any)[k]
		case int:
			cursor = cursor.([]any)[k]
		}
	}
	switch k := edit.path[len(edit.path)-1].(type) {
	case string:
		cursor.(map[string]any)[k] = edit.value
	case int:
		cursor.([]any)[k] = edit.value
	}
	result, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestNativeArithmeticReleaseRejectsChangedObservations(t *testing.T) {
	raw, source, feedback, evaluation := nativeSmokeFixture(t, "construct")
	for name, edit := range map[string]nativeSmokeEdit{
		"schema":                {[]any{"construction", "schema"}, "gooo/joint-construction/v5"},
		"space":                 {[]any{"construction", "candidate_space"}, "5"},
		"budget":                {[]any{"construction", "program_budget"}, 5},
		"selection":             {[]any{"construction", "selected_attempt"}, 4},
		"original case":         {[]any{"construction", "construction_cases", "cases", 0, "expected", "Main"}, 0},
		"initial candidate":     {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "candidate_scores", 2, "id"}, "bounded"},
		"initial rejection":     {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_fill", "rejected_candidates"}, []any{}},
		"rejected count":        {[]any{"construction", "attempts", 1, "fill_candidates", 0, "candidate_count"}, 5},
		"fault assignment":      {[]any{"construction", "attempts", 4, "fill_candidates", 0, "hole_fills", 1, "expression"}, "input.limit"},
		"fault holdout":         {[]any{"construction", "attempts", 4, "fill_candidates", 0, "holdout_cases_passed"}, 0},
		"fault runtime schema":  {[]any{"construction", "attempts", 4, "runtime", "schema"}, "gooo/body-composition-runtime/v1"},
		"fault source":          {[]any{"construction", "attempts", 4, "runtime", "original_source_sha256"}, "changed"},
		"fault count":           {[]any{"construction", "attempts", 4, "runtime", "outcomes", "faulted"}, 0},
		"missing outcome":       {[]any{"construction", "attempts", 4, "runtime", "outcomes"}, nil},
		"missing fault":         {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault"}, nil},
		"divisor":               {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "right"}, 1},
		"missing divisor":       {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "right"}, nil},
		"left operand":          {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "left"}, 7},
		"site":                  {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "site", "start"}, 0},
		"invented fault output": {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "actual"}, -8},
		"fault expected":        {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "expected"}, 8},
		"fault input":           {[]any{"construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "input", "used"}, 0},
		"native timeout":        {[]any{"construction", "attempts", 4, "runtime", "runs", 0, "timed_out"}, true},
		"native replay":         {[]any{"construction", "attempts", 4, "runtime", "runs", 1, "stdout_sha256"}, "changed"},
		"large integer":         {[]any{"evaluation", "runtime", "traces", 3, "deliveries", 0, "actual"}, json.Number("-9007199254740992")},
		"new inference":         {[]any{"evaluation", "new_model_calls"}, 1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateNativeArithmeticSmoke(changedNativeSmoke(t, raw, edit), source, feedback, evaluation, 6, false, ""); err == nil {
				t.Fatal("changed native observation accepted")
			}
		})
	}
}

func TestNativeGraphReleaseRejectsChangedOutcomes(t *testing.T) {
	raw, source, _, cases := nativeSmokeFixture(t, "graph-construct")
	var probe struct{ Runtime nativeSmokeRuntime }
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatal(err)
	}
	positions := map[string]int{}
	for i, d := range probe.Runtime.Traces[0].Deliveries {
		positions[d.ID] = i
	}
	first, divide, dependent := positions["faults://activity/first"], positions["faults://activity/divide"], positions["faults://activity/dependent"]
	for name, edit := range map[string]nativeSmokeEdit{
		"blocked count":            {[]any{"runtime", "outcomes", "blocked"}, 0},
		"missing consumer":         {[]any{"runtime", "traces", 0, "deliveries", dependent, "activity_id"}, "absent"},
		"dependency":               {[]any{"runtime", "traces", 0, "deliveries", dependent, "blocked_by"}, []string{"faults://activity/first"}},
		"invented consumer output": {[]any{"runtime", "traces", 0, "deliveries", dependent, "actual"}, 1},
		"invented producer value":  {[]any{"runtime", "traces", 0, "deliveries", dependent, "input"}, 0},
		"large integer":            {[]any{"runtime", "traces", 0, "deliveries", first, "actual"}, json.Number("9007199254740992")},
		"root input":               {[]any{"runtime", "traces", 0, "deliveries", first, "input"}, 7},
		"joined root input":        {[]any{"runtime", "traces", 0, "deliveries", divide, "inputs", 1, "value"}, 2},
		"joined producer":          {[]any{"runtime", "traces", 0, "deliveries", divide, "inputs", 0, "producer_id"}, "absent"},
		"fault operand":            {[]any{"runtime", "traces", 0, "deliveries", divide, "fault", "left"}, json.Number("9007199254740992")},
		"scored fault":             {[]any{"runtime", "traces", 0, "deliveries", divide, "passed"}, true},
		"unobserved":               {[]any{"runtime", "outcomes", "unobserved"}, 1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateNativeArithmeticGraph(changedNativeSmoke(t, raw, edit), source, cases, false, ""); err == nil {
				t.Fatal("changed graph observation accepted")
			}
		})
	}
}
