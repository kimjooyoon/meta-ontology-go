package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
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
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal("offline search did not emit a JSON report", err)
	}
	r, search := result.Report, result.Report.BodySearch
	if code != exitOK || stderr.Len() != 0 || r.Decision != "PASS" || r.Schema != "gooo/body-codegen-report/v3" ||
		r.RepositoryWrites != 0 || search == nil || search.Schema != "gooo/body-codegen-ir-search-plan/v1" ||
		search.SelectedCandidateID != "zero" || search.TrainingPassed != 5 || search.TrainingTotal != 5 ||
		search.TrainingAccuracyPercent == nil || *search.TrainingAccuracyPercent != 100 ||
		search.AttemptedCandidates != 3 || search.EvaluatedCandidates != 3 || len(search.TrainingCaseResults) != 5 ||
		len(search.Attempts) != 3 || search.Attempts[2].SelectionMethod != "sole_remaining_candidate" ||
		search.HoldoutPassed != 2 || search.HoldoutTotal != 2 || search.HoldoutAccuracyPercent == nil ||
		*search.HoldoutAccuracyPercent != 100 || search.ProviderBudgetUsedMS != 0 || search.ProviderOperations != 0 ||
		search.GlobalBestAccuracyPercent != nil || search.StopReason != "TRAINING_SUITE_PASSED" {
		t.Fatalf("offline IR body search = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if search.LocalCalls == nil || search.LocalCalls.Status != "RECORDED" || len(search.LocalCalls.Evaluations) != 5 {
		t.Fatal("CLI omitted the original evaluator history", search.LocalCalls)
	}
	for i, phase := range []string{"candidate", "candidate", "candidate", "selected-training", "selected-holdout"} {
		observed := search.LocalCalls.Evaluations[i]
		cases := 5
		if i == 4 {
			cases = 2
		}
		if observed == nil || observed.Phase != phase || observed.ActivityID != r.ActivityID ||
			observed.Status != "RECORDED" || !observed.EvaluationCompleted || observed.CasesCompleted != cases || len(observed.Calls) != 0 {
			t.Fatal("CLI evaluator phase or case count differs", observed)
		}
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
