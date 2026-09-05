package transformationeffect

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"strings"
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
		var progress bytes.Buffer
		sequence := 0
		opts := Options{ExpectedSHA: "head", ProgressWriter: &progress, InvocationID: "run/1/job:first"}
		output, runErr := runProgressPhase(opts, action, "PREFLIGHT", &sequence, func() ([]byte, error) {
			return []byte("executor-output"), nil
		})
		if runErr != nil || string(output) != "executor-output" {
			t.Fatalf("executor result changed: output=%q err=%v", output, runErr)
		}
		if sequence != 2 {
			t.Fatalf("unexpected normal sequence: %d", sequence)
		}
		events := readOperationProgress(t, progress.String())
		if len(events) != 2 || events[0].Boundary != "ENTERED" || events[1].Boundary != "RETURNED" || events[1].ReturnError != "" {
			t.Fatalf("unexpected normal progress: %#v", events)
		}
		assertProgressIdentity(t, events, opts, action)
	})

	t.Run("incomplete phase has no return", func(t *testing.T) {
		var progress bytes.Buffer
		sequence := 0
		opts := Options{ExpectedSHA: "head", ProgressWriter: &progress, InvocationID: "run/1/job:replay"}
		if err := writeOperationProgress(opts, action, "APPLY", "ENTERED", "", &sequence); err != nil {
			t.Fatal(err)
		}
		events := readOperationProgress(t, progress.String())
		if len(events) != 1 || events[0].Boundary != "ENTERED" {
			t.Fatalf("unexpected incomplete progress: %#v", events)
		}
		assertProgressIdentity(t, events, opts, action)
	})

	t.Run("diagnostic write failure preserves executor result", func(t *testing.T) {
		opts := Options{ExpectedSHA: "head", ProgressWriter: failingProgressWriter{}, InvocationID: "run/1/job:negative"}
		sequence := 0
		sentinel := errors.New("executor failure")
		called := false
		output, runErr := runProgressPhase(opts, action, "APPLY", &sequence, func() ([]byte, error) {
			called = true
			return []byte("executor-output"), sentinel
		})
		if !called || string(output) != "executor-output" || runErr != sentinel {
			t.Fatalf("diagnostic I/O changed executor result: called=%t output=%q err=%v", called, output, runErr)
		}
	})

	t.Run("different attempts remain separate", func(t *testing.T) {
		var progress bytes.Buffer
		first := Options{ExpectedSHA: "head", ProgressWriter: &progress, InvocationID: "run/1/job:first"}
		retry := Options{ExpectedSHA: "head", ProgressWriter: &progress, InvocationID: "run/2/job:first"}
		firstSequence, retrySequence := 0, 0
		if err := writeOperationProgress(first, action, "APPLY", "ENTERED", "", &firstSequence); err != nil {
			t.Fatal(err)
		}
		firstEvents := readOperationProgress(t, progress.String())
		if err := writeOperationProgress(retry, action, "APPLY", "ENTERED", "", &retrySequence); err != nil {
			t.Fatal(err)
		}
		allEvents := readOperationProgress(t, progress.String())
		if len(firstEvents) != 1 || len(allEvents) != 2 || allEvents[0].InvocationID == allEvents[1].InvocationID {
			t.Fatalf("attempt progress was combined: first=%#v all=%#v", firstEvents, allEvents)
		}
		if allEvents[0].Sequence != 1 || allEvents[1].Sequence != 1 {
			t.Fatalf("attempt sequence was not isolated: all=%#v", allEvents)
		}
	})
}

func readOperationProgress(t *testing.T, progress string) []operationProgressEvent {
	t.Helper()
	var events []operationProgressEvent
	scanner := bufio.NewScanner(strings.NewReader(progress))
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

type failingProgressWriter struct{}

func (failingProgressWriter) Write([]byte) (int, error) {
	return 0, errors.New("diagnostic writer unavailable")
}

func assertProgressIdentity(t *testing.T, events []operationProgressEvent, opts Options, action generation.Action) {
	t.Helper()
	for _, event := range events {
		if event.Schema != "gooo/transformation-effect-operation-progress/v1" || event.HeadSHA != opts.ExpectedSHA || event.InvocationID != opts.InvocationID || event.ActionIndicatorID != action.IndicatorID || event.Operation != string(action.Operation) || event.Activity != action.Activity || event.Executor != action.Executor || event.Subject != action.Subject || event.SubjectKind != string(action.SubjectKind) || event.InputContractSourceDigest != action.InputContractSourceDigest || event.InputContractSemanticDigest != action.InputContractSemanticDigest {
			t.Fatalf("progress identity mismatch: %#v", event)
		}
	}
}
