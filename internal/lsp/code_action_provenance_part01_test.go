package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func testCodeActionDigest(label string) string {
	return cache.HashBytes([]byte(label)).String()
}

func TestCodeActionProvenanceBindsSnapshotAndEnvironment(t *testing.T) {
	uri := "file:///workspace/example.gooo"
	version := 7
	sourceDigest := testCodeActionDigest("source")
	originDigest := testCodeActionDigest("origin")
	environmentDigest := testCodeActionDigest("environment")
	edits := []TextEdit{{Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 3}}, NewText: "value"}}
	observation := ObserveCodeActionProvenance(uri, version, sourceDigest, originDigest, environmentDigest, edits)
	if observation.Decision != CodeActionProvenanceClosed || !observation.NonAuthorizing {
		t.Fatalf("expected closed non-authorizing observation, got %#v", observation)
	}
	if err := ValidateCodeActionProvenance(observation); err != nil {
		t.Fatalf("expected valid observation: %v", err)
	}
	if err := ApplyCodeActionProvenance(observation, version, sourceDigest, originDigest, environmentDigest); err != nil {
		t.Fatalf("expected matching context: %v", err)
	}
	if err := ApplyCodeActionProvenance(observation, version+1, sourceDigest, originDigest, environmentDigest); err == nil {
		t.Fatal("expected document version change to reject stale edit")
	}
	if err := ApplyCodeActionProvenance(observation, version, sourceDigest, originDigest, testCodeActionDigest("other-environment")); err == nil {
		t.Fatal("expected environment change to reject stale edit")
	}
}

func TestCodeActionProvenancePreservesUnknown(t *testing.T) {
	observation := ObserveCodeActionProvenance("file:///workspace/example.gooo", 1, testCodeActionDigest("source"), testCodeActionDigest("origin"), "", []TextEdit{{NewText: "value"}})
	if observation.Decision != CodeActionProvenanceUnknown || observation.Reason != "MISSING_ENVIRONMENT_DIGEST" {
		t.Fatalf("expected unknown environment boundary, got %#v", observation)
	}
	if err := ValidateCodeActionProvenance(observation); err != nil {
		t.Fatalf("expected unknown observation to remain structurally valid: %v", err)
	}
	if err := ApplyCodeActionProvenance(observation, 1, observation.SourceDigest, observation.OriginDigest, ""); err == nil {
		t.Fatal("expected unknown observation to remain unapplied")
	}
}
