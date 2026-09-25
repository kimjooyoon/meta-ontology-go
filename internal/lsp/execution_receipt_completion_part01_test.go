package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveExecutionBackedCompletionRejectsUnboundReceipt(t *testing.T) {
	observation := ObserveExecutionBackedCompletion(
		DeclarationCompletionContext{
			DocumentURI:       "file:///workspace/example.gooo",
			DocumentVersion:   3,
			DeclarationSymbol: "activity CommitCandidate",
			EnvironmentDigest: "environment-digest",
		},
		DeclarationCompletionObservation{
			Status:            DeclarationCompletionStatusComplete,
			DocumentURI:       "file:///workspace/example.gooo",
			DocumentVersion:   3,
			DeclarationSymbol: "activity CommitCandidate",
			EnvironmentDigest: "environment-digest",
			Digest:            "completion-digest",
		},
		valueexecution.ExecutionOriginReceipt{},
	)
	if observation.Status != ExecutionBackedCompletionStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "EXECUTION_ORIGIN_UNBOUND" {
		t.Fatalf("reason=%q, want unbound receipt", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("execution-backed completion must remain non-authorizing and digestable")
	}
}
