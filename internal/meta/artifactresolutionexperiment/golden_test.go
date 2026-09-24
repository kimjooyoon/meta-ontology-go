package artifactresolutionexperiment

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func TestGoldenEqualIgnoresVolatileProvenanceDigests(t *testing.T) {
	actual := artifactemit.Artifact{
		Schema: "gooo/test/v1", Decision: "PASS", Resolution: "EXACT",
		Reason: "TEST", Kind: "test", SubjectDigest: "sha256:actual", Digest: "sha256:actual",
		Operation: artifactemit.Operation{Activity: "PayOrder"},
	}
	expected := actual
	expected.SubjectDigest = "sha256:golden"
	expected.Digest = "sha256:golden"
	if !goldenEqual(actual, expected) {
		t.Fatal("golden comparison should ignore volatile top-level provenance digests")
	}
	expected.Operation.Activity = "Other"
	if goldenEqual(actual, expected) {
		t.Fatal("golden comparison must retain semantic operation differences")
	}
}
