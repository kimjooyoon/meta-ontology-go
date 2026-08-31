package artifactemit

import "testing"

func TestValidDigestBindsArtifactContents(t *testing.T) {
	artifact := Emit(OperationManifestKind, validReceiptJSON(t))
	if !ValidDigest(artifact) {
		t.Fatal("emitted artifact digest must validate")
	}
	artifact.Operation.Activity = "Other"
	if ValidDigest(artifact) {
		t.Fatal("copied digest must not authorize changed contents")
	}
}

func TestValidSHA256RejectsPrefixOnly(t *testing.T) {
	if ValidSHA256("sha256:artifact") || ValidSHA256("sha256:") {
		t.Fatal("digest text must contain exactly 32 encoded bytes")
	}
}
