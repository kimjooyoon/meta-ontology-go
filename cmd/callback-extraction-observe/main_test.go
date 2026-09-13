package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func TestRunPreservesLocalExecutionBoundary(t *testing.T) {
	t.Setenv("CI", "")
	var stdout, stderr bytes.Buffer
	code := run([]string{"-file", "example_test.go", "-subject", "func:TestExample"}, &stdout, &stderr)
	var observation extractor.CallbackExtractionObservation
	if err := json.Unmarshal(stdout.Bytes(), &observation); err != nil {
		t.Fatal(err)
	}
	if code != 1 || stderr.Len() == 0 || observation.Decision != "UNKNOWN" ||
		observation.OperationAdmission != "UNKNOWN" || observation.ApplyPermission != "FORBIDDEN" ||
		observation.Frontier.Reason != "CI_EXECUTION_REQUIRED" || len(observation.Runs) != 0 {
		t.Fatalf("local boundary: exit=%d observation=%+v stderr=%s", code, observation, &stderr)
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"-file", "x.go"},
		{"-file", "x.go", "-subject", "func:TestX", "-timeout", "0s"},
		{"-file", "x.go", "-subject", "func:TestX", "unexpected"},
		{"-unknown"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("args=%v exit=%d stdout=%s stderr=%s", args, code, &stdout, &stderr)
		}
	}
}
