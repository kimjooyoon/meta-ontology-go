package transformationeffect

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestOperationProgressBoundaries(t *testing.T) {
	action := generation.Action{
		IndicatorID:                   "sha256:action",
		Operation:                     "extract-function",
		Activity:                      "ExtractFunction",
		Executor:                      "bootstrap/function-extractor",
		Subject:                       "fixture.go:1:Example",
		SubjectKind:                   sourcepolicy.SubjectKindFunction,
		InputContractSourceDigest:     "source-contract",
		InputContractSemanticDigest:   "semantic-contract",
	}

	t.Run("normal returned phases", func(t *testing.T) {
		progressPath := filepath.Join(t.TempDir(), "operation-progress.jsonl")
		sequence := 0
		opts := Options{ExpectedSHA: "head", ProgressPath: progressPath, InvocationID: "run/1/job:first"}
		if err := writeOperationProgress(opts, action, "PREFLIGHT", "ENTERED", "", &sequence); err != nil {
			t.Fatal(err)
		}
		if err := writeOperationProgress(opts, action, "PREFLIGHT", "RETURNED", "", &sequence); err != nil {
			t.Fatal(err)
		}
		events := readOperationProgress(t, progressPath)
		if len(events) != 2 || events[0].Boundary != "ENTERED" || events[1].Boundary != "RETURNED" || events[1].ReturnError != "" {
			t.Fatalf("unexpected normal progress: %#v", events)
		}
		assertProgressIdentity(t, events, opts, action)
	})

	t.Run("incomplete phase has no return", func(t *testing.T) {
		progressPath := filepath.Join(t.TempDir(), "operation-progress.jsonl")
		sequence := 0
		opts := Options{ExpectedSHA: "head", ProgressPath: progressPath, InvocationID: "run/1/job:replay"}
		if err := writeOperationProgress(opts, action, "APPLY", "ENTERED", "", &sequence); err != nil {
			t.Fatal(err)
		}
		events := readOperationProgress(t, progressPath)
		if len(events) != 1 || events[0].Boundary != "ENTERED" {
			t.Fatalf("unexpected incomplete progress: %#v", events)
		}
		assertProgressIdentity(t, events, opts, action)
	})

	t.Run("different attempts remain separate", func(t *testing.T) {
		root := t.TempDir()
		firstPath := filepath.Join(root, "first", "operation-progress.jsonl")
		retryPath := filepath.Join(root, "retry", "operation-progress.jsonl")
		first := Options{ExpectedSHA: "head", ProgressPath: firstPath, InvocationID: "run/1/job:first"}
		retry := Options{ExpectedSHA: "head", ProgressPath: retryPath, InvocationID: "run/2/job:first"}
		firstSequence, retrySequence := 0, 0
		if err := writeOperationProgress(first, action, "APPLY", "ENTERED", "", &firstSequence); err != nil {
			t.Fatal(err)
		}
		if err := writeOperationProgress(retry, action, "APPLY", "ENTERED", "", &retrySequence); err != nil {
			t.Fatal(err)
		}
		firstEvents := readOperationProgress(t, firstPath)
		retryEvents := readOperationProgress(t, retryPath)
		if len(firstEvents) != 1 || len(retryEvents) != 1 || firstEvents[0].InvocationID == retryEvents[0].InvocationID {
			t.Fatalf("attempt progress was combined: first=%#v retry=%#v", firstEvents, retryEvents)
		}
		if firstEvents[0].Sequence != 1 || retryEvents[0].Sequence != 1 {
			t.Fatalf("attempt sequence was not isolated: first=%#v retry=%#v", firstEvents, retryEvents)
		}
	})
}

func readOperationProgress(t *testing.T, progressPath string) []operationProgressEvent {
	t.Helper()
	file, err := os.Open(progressPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var events []operationProgressEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event operationProgressEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return events
}

func assertProgressIdentity(t *testing.T, events []operationProgressEvent, opts Options, action generation.Action) {
	t.Helper()
	for _, event := range events {
		if event.Schema != "gooo/transformation-effect-operation-progress/v1" || event.HeadSHA != opts.ExpectedSHA || event.InvocationID != opts.InvocationID || event.ActionIndicatorID != action.IndicatorID || event.Operation != string(action.Operation) || event.Activity != action.Activity || event.Executor != action.Executor || event.Subject != action.Subject || event.SubjectKind != string(action.SubjectKind) || event.InputContractSourceDigest != action.InputContractSourceDigest || event.InputContractSemanticDigest != action.InputContractSemanticDigest {
			t.Fatalf("progress identity mismatch: %#v", event)
		}
	}
}
