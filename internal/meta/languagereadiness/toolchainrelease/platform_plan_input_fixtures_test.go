package toolchainrelease

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func planInputsFixture(t *testing.T, mode string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "plan-inputs-"+mode+".json.gz"))
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
	return raw
}

func TestPlanInputReleaseKeepsUnknownSeparateFromScoredReplay(t *testing.T) {
	for _, example := range planInputExamples() {
		var plan bodyexecution.CompositionInspection
		if err := json.Unmarshal(planInputsFixture(t, example.name+"-plan"), &plan); err != nil {
			t.Fatal(err)
		}
		if err := validatePlanInputTemplate(planInputsFixture(t, example.name+"-template"), plan); err != nil {
			t.Fatal(err)
		}
		selected := ""
		for _, mode := range []string{"inputs", "inputs-replay", "scored-replay"} {
			var err error
			selected, err = validatePlanInputExecution(planInputsFixture(t, example.name+"-"+mode), plan, example, mode, selected)
			if err != nil {
				t.Fatal(example.name, mode, err)
			}
		}
	}
}

func TestPlanInputReleaseRejectsInventedScoresAndFreshReplay(t *testing.T) {
	example := planInputExamples()[0]
	var plan bodyexecution.CompositionInspection
	if err := json.Unmarshal(planInputsFixture(t, example.name+"-plan"), &plan); err != nil {
		t.Fatal(err)
	}
	raw := planInputsFixture(t, example.name+"-inputs")
	for name, edit := range map[string]nativeSmokeEdit{
		"missing total":    {[]any{"runtime", "finite_total"}, nil},
		"invented pass":    {[]any{"runtime", "finite_passed"}, 1},
		"invented oracle":  {[]any{"runtime", "traces", 0, "deliveries", 0, "passed"}, true},
		"runtime predict":  {[]any{"runtime", "model_calls"}, 1},
		"missing predict":  {[]any{"runtime", "model_calls"}, nil},
		"rounded input":    {[]any{"runtime", "traces", 0, "deliveries", 0, "inputs", 0, "value"}, 9007199254740992},
		"changed plan":     {[]any{"composition", "plan", "entry_activity"}, "Other"},
		"repeat predict":   {[]any{"runtime_history", 1, "model_calls"}, 1},
		"executable reuse": {[]any{"runtime_history", 1, "artifact", "reused"}, false},
		"changed model":    {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "model_calls"}, 0},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validatePlanInputExecution(changedNativeSmoke(t, raw, edit), plan, example, "inputs", ""); err == nil {
				t.Fatal("changed native input observation accepted")
			}
		})
	}
	replay := planInputsFixture(t, example.name+"-inputs-replay")
	if _, err := validatePlanInputExecution(changedNativeSmoke(t, replay,
		nativeSmokeEdit{[]any{"generated_now"}, true}), plan, example, "inputs-replay", ""); err == nil {
		t.Fatal("saved replay regenerated its program")
	}
	scored := planInputsFixture(t, example.name+"-scored-replay")
	for _, edit := range []nativeSmokeEdit{
		{[]any{"runtime", "traces"}, nil},
		{[]any{"runtime", "traces", 0, "deliveries", 0, "passed"}, nil},
		{[]any{"runtime", "traces", 0, "deliveries", 0, "actual", "count"}, 9007199254740992},
	} {
		if _, err := validatePlanInputExecution(changedNativeSmoke(t, scored, edit), plan, example, "scored-replay", ""); err == nil {
			t.Fatal("scored replay lost or changed its original result")
		}
	}
}
