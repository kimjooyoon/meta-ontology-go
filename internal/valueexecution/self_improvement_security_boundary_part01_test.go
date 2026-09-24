package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementSecurityBoundary(t *testing.T) {
	digest := strings.Repeat("a", 64)
	observation := ObserveSelfImprovementSecurityBoundary(
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
	if observation.Status != SelfImprovementSecurityBoundaryBound {
		t.Fatalf("status = %q, want BOUND", observation.Status)
	}
	if observation.CryptographicVerified {
		t.Fatal("observation must not claim cryptographic verification")
	}
	if !observation.NonAuthorizing {
		t.Fatal("security boundary must remain non-authorizing")
	}
	if len(observation.Digest) != 64 {
		t.Fatalf("digest length = %d, want 64", len(observation.Digest))
	}
}

func TestObserveSelfImprovementSecurityBoundaryRejectsExpiredIdentity(t *testing.T) {
	observation := ObserveSelfImprovementSecurityBoundary(
		SelfImprovementSecurityBoundaryInput{
			WorkloadID: "spiffe://example.org/workload/gooo",
			Audience:   "gooo-executor",
			Freshness:  "EXPIRED",
		},
	)
	if observation.Status != SelfImprovementSecurityBoundaryRejected {
		t.Fatalf("status = %q, want REJECTED", observation.Status)
	}
	if observation.Reason != "WORKLOAD_IDENTITY_EXPIRED" {
		t.Fatalf("reason = %q, want WORKLOAD_IDENTITY_EXPIRED", observation.Reason)
	}
}