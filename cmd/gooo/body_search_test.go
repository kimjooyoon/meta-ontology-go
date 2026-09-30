package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const bodySearchFixture = `package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity ClampNegativeToZero(Integer) -> Integer computes "let output = input; if input < 0 { output = __GOOO_BODY_HOLE_floor__ } else { output = input }; return output"
`

const validBodySearchPlan = `{
	"schema":"gooo/body-codegen-ir-search-plan/v1",
	"intent":"Clamp negative inputs to zero.",
	"hole_id":"floor",
	"candidates":[
		{"id":"identity","expression":"input"},
		{"id":"negate","expression":"-input"},
		{"id":"zero","expression":"0"}
	],
	"test_cases":[
		{"input":-2,"expected":0},
		{"input":-1,"expected":0},
		{"input":0,"expected":0},
		{"input":1,"expected":1},
		{"input":2,"expected":2}
	],
	"holdout_test_cases":[
		{"input":-9223372036854775808,"expected":0},
		{"input":9223372036854775807,"expected":9223372036854775807}
	],
	"max_attempts":3
}`

func TestRunBodyCodegenSearchWorksOfflineFromDeclaredOrder(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	reader := mapSourceReader{
		"fixture.gooo":     []byte(bodySearchFixture),
		"search-plan.json": []byte(validBodySearchPlan),
	}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(
		[]string{"--json", "--fill-search", "search-plan.json", "--activity", "ClampNegativeToZero", "fixture.gooo"},
		reader, &stdout, &stderr,
	)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"decision":"PASS"`) ||
		!strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v3"`) ||
		!strings.Contains(stdout.String(), `"body_search":{"schema":"gooo/body-codegen-ir-search-plan/v1"`) ||
		!strings.Contains(stdout.String(), `"selected_candidate_id":"zero"`) ||
		!strings.Contains(stdout.String(), `"training_passed":5`) ||
		!strings.Contains(stdout.String(), `"training_total":5`) ||
		!strings.Contains(stdout.String(), `"training_accuracy_percent":100`) ||
		!strings.Contains(stdout.String(), `"attempted_candidates":3`) ||
		!strings.Contains(stdout.String(), `"evaluated_candidates":3`) ||
		!strings.Contains(stdout.String(), `"training_case_results":[`) ||
		!strings.Contains(stdout.String(), `"selection_method":"sole_remaining_candidate"`) ||
		!strings.Contains(stdout.String(), `"holdout_passed":2`) ||
		!strings.Contains(stdout.String(), `"holdout_total":2`) ||
		!strings.Contains(stdout.String(), `"holdout_accuracy_percent":100`) ||
		!strings.Contains(stdout.String(), `"provider_budget_used_ms":0`) ||
		!strings.Contains(stdout.String(), `"global_best_accuracy_percent":null`) ||
		!strings.Contains(stdout.String(), `"stop_reason":"TRAINING_SUITE_PASSED"`) ||
		!strings.Contains(stdout.String(), `"repository_writes":0`) {
		t.Fatalf("offline IR body search = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenSearchRejectsOptionConflicts(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	reader := mapSourceReader{
		"fixture.gooo":     []byte(bodySearchFixture),
		"search-plan.json": []byte(validBodySearchPlan),
		"fill-plan.json":   []byte(`{}`),
	}
	for _, args := range [][]string{
		{"--fill-search", "search-plan.json", "--fill-plan", "fill-plan.json", "--activity", "ClampNegativeToZero", "fixture.gooo"},
		{"--fill-search", "search-plan.json", "--sample-seed", "seed", "--activity", "ClampNegativeToZero", "fixture.gooo"},
	} {
		var stdout, stderr bytes.Buffer
		code := runBodyCodegen(args, reader, &stdout, &stderr)
		if code == exitOK {
			t.Fatalf("conflicting body-codegen options were accepted: args=%q stdout=%q stderr=%q", args, stdout.String(), stderr.String())
		}
	}
}

func TestRunBodyCodegenSearchFailsClosedForMalformedPlans(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	cases := []struct {
		name string
		plan string
	}{
		{
			name: "unknown field",
			plan: strings.TrimSuffix(validBodySearchPlan, "}") + `,"unexpected":true}`,
		},
		{
			name: "missing expected value",
			plan: `{"schema":"gooo/body-codegen-ir-search-plan/v1","intent":"Clamp negative inputs.","hole_id":"floor","candidates":[{"id":"zero","expression":"0"},{"id":"identity","expression":"input"}],"test_cases":[{"input":-1}],"max_attempts":2}`,
		},
		{
			name: "all candidates invalid",
			plan: `{"schema":"gooo/body-codegen-ir-search-plan/v1","intent":"Clamp negative inputs.","hole_id":"floor","candidates":[{"id":"bool","expression":"true"},{"id":"comparison","expression":"input == 0"}],"test_cases":[{"input":-1,"expected":0}],"max_attempts":2}`,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			report := assertBodySearchPlanFailsClosed(t, test.plan)
			if test.name == "all candidates invalid" {
				assertAllInvalidBodySearchTrace(t, report)
			}
		})
	}
}

type bodySearchFailureReport struct {
	Decision   string `json:"decision"`
	Source     string `json:"source"`
	BodySearch *struct {
		TrainingAccuracyPercent *float64 `json:"training_accuracy_percent"`
		AttemptedCandidates     int      `json:"attempted_candidates"`
		EvaluatedCandidates     int      `json:"evaluated_candidates"`
		Attempts                []struct {
			ScoringCompleted bool     `json:"scoring_completed"`
			AccuracyPercent  *float64 `json:"accuracy_percent"`
		} `json:"attempts"`
	} `json:"body_search"`
}

func assertBodySearchPlanFailsClosed(t *testing.T, plan string) bodySearchFailureReport {
	t.Helper()
	reader := mapSourceReader{"fixture.gooo": []byte(bodySearchFixture), "search-plan.json": []byte(plan)}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(
		[]string{"--json", "--fill-search", "search-plan.json", "--activity", "ClampNegativeToZero", "fixture.gooo"},
		reader, &stdout, &stderr,
	)
	var report bodySearchFailureReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("search failure did not produce JSON: code=%d stdout=%q stderr=%q: %v", code, stdout.String(), stderr.String(), err)
	}
	if code != exitFailure || report.Decision != "FAIL_CLOSED" || report.Source != "" {
		t.Fatalf("invalid search plan was not closed: code=%d report=%#v stdout=%q stderr=%q", code, report, stdout.String(), stderr.String())
	}
	return report
}

func assertAllInvalidBodySearchTrace(t *testing.T, report bodySearchFailureReport) {
	t.Helper()
	if report.BodySearch == nil {
		t.Fatalf("all-invalid search did not retain its trace: %#v", report)
	}
	if report.BodySearch.AttemptedCandidates != 2 || report.BodySearch.EvaluatedCandidates != 0 ||
		len(report.BodySearch.Attempts) != 2 || report.BodySearch.TrainingAccuracyPercent != nil {
		t.Fatalf("all-invalid counts do not separate attempts from scores: %#v", report.BodySearch)
	}
	for _, attempt := range report.BodySearch.Attempts {
		if attempt.ScoringCompleted || attempt.AccuracyPercent != nil {
			t.Fatalf("unscored invalid candidate received a completed score: %#v", attempt)
		}
	}
}
