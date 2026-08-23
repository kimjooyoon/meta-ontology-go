package languageconcept

import (
	"os"
	"testing"
)

func TestArtifactBindsCatalogToMetaCodeContent(t *testing.T) {
	repository := os.DirFS("../../..")
	artifact := BuildArtifact(repository)
	if err := ValidateArtifact(repository, artifact); err != nil {
		t.Fatal(err)
	}
	if !artifact.Ready() || !artifact.ReplayEqual {
		t.Fatalf("got %s/%s replay=%v", artifact.Decision, artifact.Reason, artifact.ReplayEqual)
	}
	if artifact.Report.Summary.Concepts != 13 || artifact.Bindings.Paths != 33 {
		t.Fatalf("summary=%#v bindings=%#v", artifact.Report.Summary, artifact.Bindings)
	}
	if artifact.Bindings.Files == 0 || artifact.Bindings.Bytes == 0 || artifact.RepositoryWrites != 0 {
		t.Fatalf("effect evidence = %#v writes=%d", artifact.Bindings, artifact.RepositoryWrites)
	}
	if artifact.CatalogDigest == "" || artifact.ArtifactDigest == "" {
		t.Fatalf("missing digests: %#v", artifact)
	}
}

func TestArtifactTamperingFailsReplay(t *testing.T) {
	repository := os.DirFS("../../..")
	artifact := BuildArtifact(repository)
	artifact.Report.Concepts[0].PositiveEffect = "unsupported claim"
	artifact.ArtifactDigest = artifactDigest(artifact)
	if err := ValidateArtifact(repository, artifact); err == nil {
		t.Fatal("tampered artifact passed replay")
	}
}
