package provenance

import (
	"strings"
	"testing"
	"time"
)

func TestObserveWorkloadIdentityEvidenceBindsSVIDAndTrustBundle(t *testing.T) {
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC)
	evidence := ObserveWorkloadIdentityEvidence(observation, "sha256:"+strings.Repeat("b", 64), "sha256:"+strings.Repeat("c", 64), now)
	if evidence.Status != WorkloadIdentityEvidenceObserved || evidence.TrustDomain != "example.org" || !evidence.NonAuthorizing || evidence.EvidenceDigest == "" {
		t.Fatalf("unexpected observed evidence: %#v", evidence)
	}
	if expired := ObserveWorkloadIdentityEvidence(observation, evidence.SVIDDigest, evidence.TrustBundleDigest, now.Add(2*time.Second)); expired.Status != WorkloadIdentityEvidenceExpired {
		t.Fatalf("expected expired evidence, got %#v", expired)
	}
}

func TestObserveWorkloadIdentityEvidencePreservesUnknownBundle(t *testing.T) {
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := ObserveWorkloadIdentityEvidence(observation, "", "sha256:"+strings.Repeat("c", 64), time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC))
	if evidence.Status != WorkloadIdentityEvidenceUnknown || evidence.Reason != "MISSING_SVID_DIGEST" {
		t.Fatalf("expected unknown SVID evidence, got %#v", evidence)
	}
}
