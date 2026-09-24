package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementReverseClosure(t *testing.T) {
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
	observation := ObserveSelfImprovementReverseClosure(
		SelfImprovementReverseClosureInput{
			ForwardDigest:     "forward-digest",
			Outcome:           "COMPLETED",
			OutcomeDigest:     "outcome-digest",
			SecurityBoundary: security,
		},
	)
	if observation.Status != SelfImprovementReverseClosureCompleted {
		t.Fatalf("status = %q, want COMPLETED", observation.Status)
	}
	if !observation.ReverseObservation {
		t.Fatal("reverse observation must be true")
	}
	if !observation.NonAuthorizing {
		t.Fatal("reverse closure must remain non-authorizing")
	}
	if len(observation.Digest) != 64 {
		t.Fatalf("digest length = %d, want 64", len(observation.Digest))
	}
}

func TestObserveSelfImprovementReverseClosurePreservesRejection(t *testing.T) {
	observation := ObserveSelfImprovementReverseClosure(
		SelfImprovementReverseClosureInput{
			ForwardDigest:     "forward-digest",
			Outcome:           "COMPLETED",
			OutcomeDigest:     "outcome-digest",
			SecurityBoundary: SelfImprovementExecutionSecurityBoundaryObservation{
				Status: SelfImprovementSecurityBoundaryRejected,
				Digest: "boundary-digest",
			},
		},
	)
	if observation.Status != SelfImprovementReverseClosureNotApplied {
		t.Fatalf("status = %q, want NOT_APPLIED", observation.Status)
	}
	if observation.Reason != "SECURITY_BOUNDARY_REJECTED" {
		t.Fatalf("reason = %q, want SECURITY_BOUNDARY_REJECTED", observation.Reason)
	}
}