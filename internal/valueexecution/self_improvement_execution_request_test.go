package valueexecution

import "testing"

func TestObserveSelfImprovementExecutionRequestFailsClosed(t *testing.T) {
	observation := ObserveSelfImprovementExecutionRequest(
		SelfImprovementCandidateSelectionObservation{},
		SelfImprovementExecutionRequestContext{},
	)
	if observation.Status != SelfImprovementExecutionRequestStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "EXECUTION_REQUEST_SELECTION_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted selection", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("execution request must remain non-authorizing and digestable")
	}
}

func TestObserveSelfImprovementExecutionRequestRejectsMalformedSecurityDigest(t *testing.T) {
	selection := SelfImprovementCandidateSelectionObservation{
		Schema:         SelfImprovementCandidateSelectionObservationSchema,
		Status:         SelfImprovementCandidateSelectionStatusSelected,
		NonAuthorizing: true,
		Digest:         "sha256:0000000000000000000000000000000000000000000000000000000000000000",
	}
	observation := ObserveSelfImprovementExecutionRequest(
		selection,
		SelfImprovementExecutionRequestContext{
			RequestID:              "request-1",
			SourceURI:              "file:///workspace/candidate.gooo",
			EnvironmentDigest:      "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			WorkloadIdentityDigest: "not-a-digest",
			GatewayPolicyDigest:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		},
	)
	if observation.Status != SelfImprovementExecutionRequestStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "EXECUTION_REQUEST_SECURITY_DIGEST_INVALID" {
		t.Fatalf("reason=%q, want invalid security digest", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("malformed security digest must remain non-authorizing and digestable")
	}
}
