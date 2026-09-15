package selfimprovementattestation

import "testing"

func TestResolveFailsClosedOnWorkflowSHAMismatch(t *testing.T) {
	request := validRequest()
	request.Verification[0].VerificationResult.Signature.Certificate.GitHubWorkflowSHA = "wrong"
	receipt, err := Resolve(request)
	if err == nil {
		t.Fatal("expected workflow SHA mismatch to fail closed")
	}
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "PRODUCER_WORKFLOW_SHA_MISMATCH" {
		t.Fatalf("unexpected failure: decision=%s reason=%s", receipt.Decision, receipt.Reason)
	}
	if receipt.Metrics.FalseTotal != 1 || receipt.Metrics.FalsePromotionCount != 0 {
		t.Fatalf("mismatch metrics were promoted: %+v", receipt.Metrics)
	}
}

func TestResolveFailsClosedOnArchiveDigestMismatch(t *testing.T) {
	request := validRequest()
	request.ArchiveDigest = "sha256:mismatch"
	receipt, err := Resolve(request)
	if err == nil {
		t.Fatal("expected archive mismatch to fail closed")
	}
	if receipt.Reason != "ATTESTED_ARCHIVE_DIGEST_MISMATCH" {
		t.Fatalf("unexpected reason %q", receipt.Reason)
	}
}
