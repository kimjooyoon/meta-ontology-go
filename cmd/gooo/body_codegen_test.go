package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRunBodyCodegenWritesProjectionOrClosedFailure(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	valid := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Clamp(Integer) -> Integer computes "if input > 0 { return input } else { return 0 }"
`
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--activity", "Clamp", "fixture.gooo"}, fixtureReader{source: valid}, &stdout, &stderr)
	if code != exitOK || !strings.Contains(stdout.String(), "func Clamp(input int64) int64") || stderr.Len() != 0 {
		t.Fatalf("body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runBodyCodegen([]string{"--json", "--activity", "Clamp", "fixture.gooo"}, fixtureReader{source: strings.Replace(valid, "return input", "return helper(input)", 1)}, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v3"`) || !strings.Contains(stdout.String(), `"decision":"FAIL_CLOSED"`) || !strings.Contains(stdout.String(), `"repository_writes":0`) || !strings.Contains(stdout.String(), `"completeness_receipt":{"schema":"gooo/metaprogramming-completeness-receipt/v2"`) || !strings.Contains(stdout.String(), `"fail_closed_reason":"`) {
		t.Fatalf("failed body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenInfersIntegerLiteralLocalAsInt64(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity ConstantLocal(Integer) -> Integer computes "let result = 5\nreturn result"
`
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--activity", "ConstantLocal", "fixture.gooo"}, fixtureReader{source: fixture}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), "var result int64 = 5") {
		t.Fatalf("integer-local body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenFillsTypedIRHoleAndReportsFunctionalScore(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Clamp(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }"
`
	plan := `{
		"schema":"gooo/body-codegen-ir-fill-plan/v1",
		"intent":"Clamp negative inputs to zero.",
		"hole_id":"floor",
		"candidates":[{"id":"zero","expression":"0"},{"id":"identity","expression":"input"}],
		"test_cases":[{"input":-1,"expected":0},{"input":0,"expected":0},{"input":1,"expected":1}]
	}`
	reader := mapSourceReader{
		"fixture.gooo":   []byte(fixture),
		"fill-plan.json": []byte(plan),
	}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(
		[]string{"--json", "--fill-plan", "fill-plan.json", "--activity", "Clamp", "fixture.gooo"},
		reader, &stdout, &stderr,
	)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"selected_candidate_id":"zero"`) ||
		!strings.Contains(stdout.String(), `"functional_accuracy_percent":100`) ||
		!strings.Contains(stdout.String(), `"test_cases_passed":3`) ||
		!strings.Contains(stdout.String(), `"execution_model":"synchronous_sequential_no_background_codegen_goroutines"`) {
		t.Fatalf("IR body-fill = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenFillsSeveralTypedIRHolesFromOneDeclaredAssignment(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Lift(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + `
`
	plan := `{
		"schema":"gooo/body-codegen-ir-fill-plan/v2",
		"intent":"Add one to the input by choosing the base and increment together.",
		"holes":[{"id":"seed"},{"id":"step"}],
		"candidates":[
			{"id":"compose","fills":{"seed":"input + 0","step":"1"}},
			{"id":"double","fills":{"seed":"input * 2","step":"0"}},
			{"id":"subtract","fills":{"seed":"input - 0","step":"-1"}}
		],
		"test_cases":[{"input":0,"expected":1},{"input":1,"expected":2},{"input":2,"expected":3}]
	}`
	reader := mapSourceReader{"fixture.gooo": []byte(fixture), "fill-plan.json": []byte(plan)}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(
		[]string{"--json", "--fill-plan", "fill-plan.json", "--activity", "Lift", "fixture.gooo"},
		reader, &stdout, &stderr,
	)
	var report struct {
		Report struct {
			BodyFill struct {
				Schema                string              `json:"schema"`
				SelectedCandidateID   string              `json:"selected_candidate_id"`
				FunctionalAccuracyPct float64             `json:"functional_accuracy_percent"`
				HoleFills             []map[string]string `json:"hole_fills"`
			} `json:"body_fill"`
		} `json:"report"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("body-codegen did not return JSON: code=%d stdout=%q stderr=%q: %v", code, stdout.String(), stderr.String(), err)
	}
	if code != exitOK || stderr.Len() != 0 || report.Report.BodyFill.Schema != "gooo/body-codegen-ir-fill-plan/v2" ||
		report.Report.BodyFill.SelectedCandidateID != "compose" || report.Report.BodyFill.FunctionalAccuracyPct != 100 ||
		len(report.Report.BodyFill.HoleFills) != 2 || !strings.Contains(report.Source, "var base = (input + 0)") ||
		!strings.Contains(report.Source, "var increment int64 = 1") || strings.Contains(report.Source, "__GOOO_BODY_HOLE_") {
		t.Fatalf("multi-hole IR fill did not emit the full selected assignment: code=%d report=%#v source=%q stderr=%q", code, report.Report.BodyFill, report.Source, stderr.String())
	}
}

func TestRunBodyCodegenFailsClosedForInvalidIRBodyFill(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Clamp(Integer) -> Integer computes "if input < 0 { return __GOOO_BODY_HOLE_floor__ } else { return input }"
`
	cases := []struct {
		name string
		plan string
	}{
		{
			name: "ill-typed candidate",
			plan: `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Clamp negative inputs.","hole_id":"floor","candidates":[{"id":"bad","expression":"true"},{"id":"zero","expression":"0"}],"test_cases":[{"input":-1,"expected":0}]}`,
		},
		{
			name: "missing expected value",
			plan: `{"schema":"gooo/body-codegen-ir-fill-plan/v1","intent":"Clamp negative inputs.","hole_id":"floor","candidates":[{"id":"zero","expression":"0"},{"id":"identity","expression":"input"}],"test_cases":[{"input":-1}]}`,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			reader := mapSourceReader{
				"fixture.gooo":   []byte(fixture),
				"fill-plan.json": []byte(test.plan),
			}
			var stdout, stderr bytes.Buffer
			code := runBodyCodegen([]string{"--json", "--fill-plan", "fill-plan.json", "--activity", "Clamp", "fixture.gooo"}, reader, &stdout, &stderr)
			var report struct {
				Decision string `json:"decision"`
				Source   string `json:"source"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("body-codegen did not return a failure report: code=%d stdout=%q stderr=%q: %v", code, stdout.String(), stderr.String(), err)
			}
			if code != exitFailure || report.Decision != "FAIL_CLOSED" || report.Source != "" || stdout.Len() == 0 {
				t.Fatalf("invalid IR body fill was not a closed failure: code=%d report=%#v stdout=%q stderr=%q", code, report, stdout.String(), stderr.String())
			}
		})
	}
}

type mapSourceReader map[string][]byte

func (r mapSourceReader) ReadFile(path string) ([]byte, error) {
	content, ok := r[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return content, nil
}

func TestRunBodyCodegenAcceptsReplayableSampleSeed(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	valid := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 0 { return input } else { return 0 }"
`
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--sample-seed", "cli-replay-seed", "--activity", "Choose", "fixture.gooo"}, fixtureReader{source: valid}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v3"`) ||
		!strings.Contains(stdout.String(), `"schema":"gooo/metaprogramming-completeness-receipt/v2"`) ||
		!strings.Contains(stdout.String(), `"aggregate_completeness_score":null`) ||
		!strings.Contains(stdout.String(), `"method":"sha256_seeded_weighted_choice/v1"`) ||
		!strings.Contains(stdout.String(), `"seed_sha256":"sha256:`) {
		t.Fatalf("seeded body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "cli-replay-seed") {
		t.Fatalf("raw sample seed leaked into JSON output: %s", stdout.String())
	}
}
