package toolchainrelease

import (
	"encoding/json"
	"testing"
)

func TestPackageCallerReleaseInputObservationAndLaterEvaluation(t *testing.T) {
	reference := packageConstructionTestReference(t)
	for _, step := range []packageConstructionSmokeRun{
		{"partial-inputs", 5, true, ""}, {"inputs", 6, true, ""},
		{"inputs-replay", 6, true, "construct"}, {"inputs-again", 6, true, "inputs"},
		{"inputs-evaluate", 6, false, "inputs-again"},
	} {
		t.Run(step.mode, func(t *testing.T) {
			var saved []byte
			if step.from != "" {
				saved = packageConstructionFixture(t, step.from)
			}
			raw := packageConstructionFixture(t, step.mode)
			if err := validatePackageConstructionRun(raw, reference, step.budget, saved, step.inputOnly); err != nil {
				t.Fatal(err)
			}
			if err := validatePackageConstructionRun(raw, reference, step.budget, saved, !step.inputOnly); err == nil {
				t.Fatal("input observation and labelled evaluation were interchangeable")
			}
		})
	}
}

func TestPackageCallerReleaseRejectsInventedInputEvidence(t *testing.T) {
	reference := packageConstructionTestReference(t)
	original := packageConstructionFixture(t, "inputs")
	for name, edit := range map[string]nativeSmokeEdit{
		"input digest":       {[]any{"inputs_digest"}, "sha256:changed"},
		"invented cases":     {[]any{"cases_digest"}, packageSmokeSHA(reference.Evaluation)},
		"scored decision":    {[]any{"decision"}, "COMPLETE_FINITE"},
		"runtime schema":     {[]any{"result", "evaluation", "runtime", "schema"}, "changed"},
		"numerator":          {[]any{"result", "evaluation", "runtime", "finite_passed"}, json.Number("4")},
		"denominator":        {[]any{"result", "evaluation", "runtime", "finite_total"}, json.Number("4")},
		"runtime inference":  {[]any{"result", "evaluation", "runtime", "model_calls"}, json.Number("1")},
		"case order":         {[]any{"result", "evaluation", "runtime", "traces", 0, "case_index"}, json.Number("1")},
		"changed input":      {[]any{"result", "evaluation", "runtime", "traces", 2, "deliveries", 0, "input", "used"}, json.Number("9007199254740993")},
		"rounded output":     {[]any{"result", "evaluation", "runtime", "traces", 2, "deliveries", 0, "actual"}, json.Number("-9007199254740992")},
		"invented expected":  {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "expected"}, json.Number("3")},
		"invented pass":      {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "passed"}, true},
		"invented fail":      {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "passed"}, false},
		"wrong entry":        {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "activity_id"}, packageBudgetID},
		"faulted entry":      {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "fault"}, map[string]any{"kind": "ZERO_DIVISOR"}},
		"blocked entry":      {[]any{"result", "evaluation", "runtime", "traces", 0, "deliveries", 0, "blocked_by"}, []string{packageBudgetID}},
		"incomplete process": {[]any{"result", "evaluation", "runtime", "runs", 0, "completed"}, false},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validatePackageConstructionRun(changedNativeSmoke(t, original, edit), reference, 6, nil, true); err == nil {
				t.Fatal("changed input evidence accepted")
			}
		})
	}
	reference.Inputs = append(reference.Inputs, '\n')
	if err := validatePackageConstructionRun(original, reference, 6, nil, true); err == nil {
		t.Fatal("unbound original input file accepted")
	}
}

func TestPackageCallerReleaseInputReplayKeepsOriginalHistory(t *testing.T) {
	reference := packageConstructionTestReference(t)
	saved := packageConstructionFixture(t, "inputs")
	replay := packageConstructionFixture(t, "inputs-again")
	for _, edit := range []nativeSmokeEdit{
		{[]any{"replayed_from_sha256"}, "sha256:changed"},
		{[]any{"result", "replayed_from_sha256"}, "sha256:changed"},
		{[]any{"result", "construction", "elapsed_ns"}, json.Number("1")},
		{[]any{"result", "evaluation", "runtime", "schema"}, "gooo/body-composition-runtime/v1"},
	} {
		if err := validatePackageConstructionRun(changedNativeSmoke(t, replay, edit), reference, 6, saved, true); err == nil {
			t.Fatal("changed saved input observation accepted")
		}
	}
	if err := validatePackageConstructionRun(replay, reference, 6, nil, true); err == nil {
		t.Fatal("unbound input replay accepted")
	}
}
