package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementExecutionSecurityBoundary(t *testing.T) {
	digest := strings.Repeat("a", 64)
	security := ObserveSelfImprovementSecurityBoundary(
		SelfImprovementSecurityBoundaryInput{
			WorkloadID:             "spiffe://example.org/workload/gooo",
			Audience:               "gooo-executor",
			Freshness:              "FRESH",
			AttestationDigest:      digest,
			SVIDDigest:             digest,
			TrustBundleDigest:      digest,
			WorkloadIdentityDigest: digest,
		},
	)
	observation := ObserveSelfImprovementExecutionSecurityBoundary(
		SelfImprovementExecutionSecurityBoundaryInput{
			RequestID:        "request-1",
			RequestDigest:    "request-digest",
			SecurityBoundary: security,
		},
	)
	if observation.Status != SelfImprovementExecutionSecurityBoundaryReady {
		t.Fatalf("status = %q, want READY", observation.Status)
	}
	if !observation.NonAuthorizing {
		t.Fatal("execution security observation must remain non-authorizing")
	}
	if len(observation.Digest) != 64 {
		t.Fatalf("digest length = %d, want 64", len(observation.Digest))
	}
}

func TestObserveSelfImprovementExecutionSecurityBoundaryRejectsBoundary(t *testing.T) {
	observation := ObserveSelfImprovementExecutionSecurityBoundary(
		SelfImprovementExecutionSecurityBoundaryInput{
			RequestID: "request-1",
			SecurityBoundary: SelfImprovementSecurityBoundaryObservation{
				Status: SelfImprovementSecurityBoundaryRejected,
				Digest: "boundary-digest",
			},
		},
	)
	if observation.Status != SelfImprovementExecutionSecurityBoundaryUnknown {
		t.Fatalf("status = %q, want UNKNOWN for missing request digest", observation.Status)
	}
	if observation.Reason != "EXECUTION_REQUEST_IDENTITY_MISSING" {
		t.Fatalf("reason = %q, want EXECUTION_REQUEST_IDENTITY_MISSING", observation.Reason)
	}
}