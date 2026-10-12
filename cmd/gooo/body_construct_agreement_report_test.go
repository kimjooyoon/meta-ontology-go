package main

import (
	"bytes"
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func agreementAttempt(local, caller bool) bodyexecution.JointAttempt {
	a := bodyexecution.JointAttempt{CriteriaTotal: 1,
		Runtime: bodyexecution.CompositionRuntime{Stage: "COMPLETE", FiniteTotal: 1}}
	status := "UNSATISFIED"
	if local {
		a.CriteriaPassed, status = 1, "SATISFIED"
	} else {
		a.CriteriaFailed = 1
	}
	if caller {
		a.Runtime.FinitePassed = 1
	}
	actual, _ := json.Marshal(caller)
	a.Runtime.Traces = []bodyexecution.CompositionTrace{{CaseIndex: 3,
		Deliveries: []bodyexecution.CompositionDelivery{{ActivityID: "relation://nondecreasing",
			Actual: actual, Expected: json.RawMessage(`true`), Passed: &caller}}}}
	a.Candidates = []bodycodegen.RecordCandidate{{Activity: "Fee", Attempt: bodycodegen.RecordAssemblyAttempt{
		Criteria: &bodycodegen.RecordCriteriaEvidence{Activity: "Fee", ActivityID: "relation://fee",
			Inputs: []bodycodegen.RecordCriteriaInput{{InputJSON: "[4]", ActualJSON: "-7",
				Checks: []bodycodegen.RecordCriterionCheck{{ID: "decreasing", Status: status}}}}}}}}
	return a
}

func TestConstructionAgreementRecordedMatrixAndWitnesses(t *testing.T) {
	c := bodyexecution.JointConstruction{CandidateSpace: "999999999999999999999999", SelectedAttempt: 0,
		Attempts: []bodyexecution.JointAttempt{agreementAttempt(true, false), agreementAttempt(false, true),
			agreementAttempt(true, true), agreementAttempt(false, false), {}}}
	before, _ := json.Marshal(c)
	rows := constructionAgreementRows(c)
	if len(rows) != 4 || rows[0][1] != "1 both satisfied; 1 local only; 1 caller only; 1 neither; 1 unclassified (missing, unavailable or inconsistent observations)" {
		t.Fatal(rows)
	}
	for _, want := range []string{"attempt 1 satisfies local", "case index 3", "actual false; expected true"} {
		if !strings.Contains(rows[1][1], want) {
			t.Fatal(want, rows[1])
		}
	}
	for _, want := range []string{"attempt 2 matches caller", "Fee (relation://fee)", "criterion decreasing returned false", "input [4]; actual -7"} {
		if !strings.Contains(rows[2][1], want) {
			t.Fatal(want, rows[2])
		}
	}
	after, _ := json.Marshal(c)
	if !bytes.Equal(before, after) || !strings.Contains(rows[3][1], "untried choices") {
		t.Fatal("report changed history or claimed declared-space closure")
	}
	if constructionAgreementRows(bodyexecution.JointConstruction{}) != nil {
		t.Fatal("missing attempts were presented as observed zeros")
	}
}

func TestConstructionAgreementUnavailableAndInconsistentRemainUnclassified(t *testing.T) {
	for name, change := range map[string]func(*bodyexecution.JointAttempt){
		"rejected":                func(a *bodyexecution.JointAttempt) { a.Rejection = &bodyexecution.JointCandidateRejection{} },
		"unknown criterion":       func(a *bodyexecution.JointAttempt) { a.CriteriaPassed, a.CriteriaUnknown = 0, 1 },
		"missing criterion count": func(a *bodyexecution.JointAttempt) { a.CriteriaPassed = 0 },
		"no local checks":         func(a *bodyexecution.JointAttempt) { a.CriteriaPassed, a.CriteriaTotal = 0, 0 },
		"negative counts":         func(a *bodyexecution.JointAttempt) { a.LocalPassed = -1 },
		"too many successes":      func(a *bodyexecution.JointAttempt) { a.CriteriaPassed = 2 },
		"missing caller":          func(a *bodyexecution.JointAttempt) { a.Runtime.Traces = nil },
		"no expectations":         func(a *bodyexecution.JointAttempt) { a.Runtime.FiniteTotal = 0 },
		"caller mismatch count":   func(a *bodyexecution.JointAttempt) { a.Runtime.FinitePassed = 0 },
		"tool failure":            func(a *bodyexecution.JointAttempt) { a.Runtime.Stage = "TOOLCHAIN" },
		"recorded failure":        func(a *bodyexecution.JointAttempt) { a.Runtime.Failure = "stopped" },
		"fault": func(a *bodyexecution.JointAttempt) {
			a.Runtime.Traces[0].Deliveries[0].Fault = &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}
		},
		"unscored fault": func(a *bodyexecution.JointAttempt) {
			a.Runtime.Traces[0].Deliveries = append(a.Runtime.Traces[0].Deliveries,
				bodyexecution.CompositionDelivery{Fault: &bodyexecution.CompositionFault{Kind: "ZERO_DIVISOR"}})
		},
		"blocked": func(a *bodyexecution.JointAttempt) {
			a.Runtime.Traces[0].Deliveries[0].BlockedBy = []string{"producer"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			a := agreementAttempt(true, true)
			change(&a)
			rows := constructionAgreementRows(bodyexecution.JointConstruction{Attempts: []bodyexecution.JointAttempt{a}})
			if len(rows) != 2 || !strings.Contains(rows[0][1], "0 neither; 1 unclassified") {
				t.Fatal(rows)
			}
		})
	}
}

func TestConstructionAgreementExactCaseFallback(t *testing.T) {
	a := agreementAttempt(false, true)
	a.CriteriaTotal, a.CriteriaFailed, a.LocalTotal, a.LocalPassed, a.Candidates = 0, 0, 2, 1, nil
	rows := constructionAgreementRows(bodyexecution.JointConstruction{Attempts: []bodyexecution.JointAttempt{a}})
	if !strings.Contains(rows[0][1], "1 caller only") || !strings.Contains(rows[1][1], "local checks 1/2") {
		t.Fatal(rows)
	}
}

func TestConstructionAgreementSavedReportAddsNoExecutionOrJSONChanges(t *testing.T) {
	saved := bodyConstructOutput{GeneratedNow: true, Construction: bodyexecution.JointConstruction{
		Decision: "PARTIAL_FINITE", StopReason: "DECLARED_SPACE_EXHAUSTED", SelectedAttempt: 0,
		Attempts: []bodyexecution.JointAttempt{agreementAttempt(true, false), agreementAttempt(false, true)}}}
	raw, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "observation.json")
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	for _, format := range []string{"text", "markdown", "json"} {
		var out, diagnostics bytes.Buffer
		if code := run([]string{"body-construct", "--report", file, "--format", format}, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String())
		}
		if format == "json" {
			if !bytes.Equal(raw, out.Bytes()) {
				t.Fatal("saved JSON changed")
			}
			continue
		}
		for _, want := range []string{"0 both satisfied; 1 local only; 1 caller only", "criterion decreasing returned false",
			"0 model calls and 0 program executions", "Recorded claims have not been reverified", "PARTIAL_FINITE"} {
			if !strings.Contains(html.UnescapeString(out.String()), want) {
				t.Fatal(want, out.String())
			}
		}
	}
	after, _ := os.ReadFile(file)
	if !bytes.Equal(raw, after) {
		t.Fatal("report changed the input file")
	}
}
