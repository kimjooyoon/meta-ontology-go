package transformationeffect

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayDiagnosticPreservesLastEnteredOperation(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "ledger.json")
	progress := []operationProgressEvent{
		{Schema: "gooo/transformation-effect-operation-progress/v1", HeadSHA: "head", InvocationID: "run/1/job/token:ledger", Sequence: 1, ActionIndicatorID: "sha256:action", Operation: "extract-function", Activity: "ExtractFunction", Executor: "bootstrap/function-extractor", Subject: "fixture.go:1:Example", SubjectKind: "function", Phase: "PREFLIGHT", Boundary: "ENTERED"},
		{Schema: "gooo/transformation-effect-operation-progress/v1", HeadSHA: "head", InvocationID: "run/1/job/token:ledger", Sequence: 2, ActionIndicatorID: "sha256:action", Operation: "extract-function", Activity: "ExtractFunction", Executor: "bootstrap/function-extractor", Subject: "fixture.go:1:Example", SubjectKind: "function", Phase: "PREFLIGHT", Boundary: "RETURNED"},
		{Schema: "gooo/transformation-effect-operation-progress/v1", HeadSHA: "head", InvocationID: "run/1/job/token:ledger", Sequence: 3, ActionIndicatorID: "sha256:action", Operation: "extract-function", Activity: "ExtractFunction", Executor: "bootstrap/function-extractor", Subject: "fixture.go:1:Example", SubjectKind: "function", Phase: "APPLY", Boundary: "ENTERED"},
	}
	file, err := os.Create(filepath.Join(directory, OperationProgressFilename))
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	for _, event := range progress {
		if err := encoder.Encode(event); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := WriteReplayDiagnostic(output, errors.New("interrupted")); err != nil {
		t.Fatal(err)
	}
	diagnostic := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if diagnostic.Decision != "UNKNOWN" || diagnostic.NextOperation != "observe-interrupted-operation" || diagnostic.ActiveOperation == nil {
		t.Fatalf("interrupted operation was not preserved: %#v", diagnostic)
	}
	if diagnostic.ActiveOperation.Operation != "extract-function" || diagnostic.ActiveOperation.Phase != "APPLY" || diagnostic.ActiveOperation.Boundary != "ENTERED" {
		t.Fatalf("active operation = %#v", diagnostic.ActiveOperation)
	}
}

func TestReplayDiagnosticDoesNotTreatReturnedProgressAsActive(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "ledger.json")
	progress := operationProgressEvent{Schema: "gooo/transformation-effect-operation-progress/v1", InvocationID: "run/1/job/token:ledger", Sequence: 1, ActionIndicatorID: "sha256:action", Phase: "APPLY", Boundary: "RETURNED"}
	payload, err := json.Marshal(progress)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, OperationProgressFilename), append(payload, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteReplayDiagnostic(output, errors.New("uncataloged")); err != nil {
		t.Fatal(err)
	}
	diagnostic := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if diagnostic.ActiveOperation != nil || diagnostic.NextOperation != "report-counterexample" {
		t.Fatalf("returned operation was treated as active: %#v", diagnostic)
	}
}
