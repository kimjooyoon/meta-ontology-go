package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestObserveDeclarationCompletionRejectsMissingOrigin(t *testing.T) {
	observation := ObserveDeclarationCompletion(
		DeclarationCompletionContext{
			DocumentURI:       "file:///workspace/example.gooo",
			DocumentVersion:   3,
			DeclarationSymbol: "activity CommitCandidate",
			EnvironmentDigest: "environment-digest",
		},
		"CommitCandidate",
		"Integer -> Integer",
		provenance.OriginChainObservation{},
	)
	if observation.Status != DeclarationCompletionStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "ORIGIN_OBSERVATION_MISSING" {
		t.Fatalf("reason=%q, want missing origin", observation.Reason)
	}
	if !observation.NonAuthorizing {
		t.Fatal("completion observation must remain non-authorizing")
	}
	if observation.Digest == "" || observation.OriginDigest == "" {
		t.Fatal("completion observation must retain provenance digests")
	}
}
