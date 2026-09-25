package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestRunSourceContinuationConsumesDeclaredFeedback(t *testing.T) {
	source := []byte(`package budget
namespace budget
entity Integer id "budget://entity/integer"
activity Next(Integer) -> Integer computes "int.add:-1"
feedback Next.result -> Next.input
`)
	reader := runSourceReaderWithFiles{"budget.gooo": source, "input.json": []byte(`{"value":3}`)}
	var stdout, stderr bytes.Buffer
	code := runSource([]string{"--json", "--entry", "Next", "--input", "input.json", "--iterations", "2", "budget.gooo"}, reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
	var report struct {
		Decision     string                      `json:"decision"`
		Continuation valueexecution.Continuation `json:"continuation"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	trace := report.Continuation
	if report.Decision != "PASS" || trace.IterationsCompleted != 2 || trace.FeedbackDeliveries != 1 || len(trace.Executions) != 2 {
		t.Fatalf("continuation=%+v", report)
	}
	if trace.Executions[0].Results["Next"].Value != 2 || trace.Executions[1].Results["Next"].Value != 1 {
		t.Fatalf("feedback was not executed: %+v", trace.Executions)
	}
}

func TestRunSourceContinuationRejectsUnsupportedInvocation(t *testing.T) {
	cases := [][]string{
		{"--entry", "Next", "--iterations", "2", "budget.gooo"},
		{"--entry", "Next", "--input", "input.json", "--iterations", "0", "budget.gooo"},
		{"--entry", "Next", "--input", "input.json", "--iterations", "-1", "budget.gooo"},
		{"--entry", "Next", "--input", "input.json", "--iterations", "2", "--iterations", "3", "budget.gooo"},
		{"--entry", "Next", "--record-input", "input.json", "--iterations", "2", "budget.gooo"},
	}
	for index, args := range cases {
		if _, err := parseRunSourceArguments(args); err == nil {
			t.Fatalf("unsupported continuation arguments %d accepted", index)
		}
	}
}
