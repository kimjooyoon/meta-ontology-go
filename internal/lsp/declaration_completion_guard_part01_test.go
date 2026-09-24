package lsp

import "testing"

func TestObserveDeclarationCompletionGuardRejectsStaleVersion(t *testing.T) {
	proposed := DeclarationCompletionContext{
		DocumentURI:       "file:///workspace/example.gooo",
		DocumentVersion:   3,
		DeclarationSymbol: "activity CommitCandidate",
		EnvironmentDigest: "environment-digest",
	}
	current := proposed
	current.DocumentVersion = 4
	observation := ObserveDeclarationCompletionGuard(proposed, current)
	if observation.Status != DeclarationCompletionGuardStatusStale {
		t.Fatalf("status=%q, want stale", observation.Status)
	}
	if observation.Reason != "COMPLETION_CONTEXT_STALE" {
		t.Fatalf("reason=%q, want stale context", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("stale guard must remain non-authorizing and digestable")
	}
}
