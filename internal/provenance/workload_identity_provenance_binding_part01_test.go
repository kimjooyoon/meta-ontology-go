package provenance

import (
	"strings"
	"testing"
	"time"
)

func TestBindWorkloadIdentityEvidenceToCompleteProvenanceIsObserved(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01(
		"sha256:"+strings.Repeat("1", 64),
		"sha256:"+strings.Repeat("2", 64),
		"sha256:"+strings.Repeat("3", 64),
		"sha256:"+strings.Repeat("4", 64),
		"sha256:"+strings.Repeat("5", 64),
		"sha256:"+strings.Repeat("6", 64),
	)
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := ObserveWorkloadIdentityEvidence(
		observation,
		"sha256:"+strings.Repeat("b", 64),
		"sha256:"+strings.Repeat("c", 64),
		time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC),
	)
	binding := BindWorkloadIdentityEvidenceToProvenancePart01(chain, evidence)
	if binding.Status != WorkloadIdentityProvenanceBindingObserved || binding.Reason != "PROVENANCE_AND_IDENTITY_EVIDENCE_BOUND" {
		t.Fatalf("unexpected observed binding: %#v", binding)
	}
	if binding.ProvenanceChainDigest != chain.ChainDigest || binding.IdentityEvidenceDigest != evidence.EvidenceDigest || !binding.NonAuthorizing {
		t.Fatalf("binding lost evidence identities: %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBindWorkloadIdentityEvidenceKeepsIncompleteProvenanceUnknown(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01(
		"sha256:"+strings.Repeat("1", 64),
		"sha256:"+strings.Repeat("2", 64),
		"sha256:"+strings.Repeat("3", 64),
		"sha256:"+strings.Repeat("4", 64),
		"",
		"",
	)
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := ObserveWorkloadIdentityEvidence(
		observation,
		"sha256:"+strings.Repeat("b", 64),
		"sha256:"+strings.Repeat("c", 64),
		time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC),
	)
	binding := BindWorkloadIdentityEvidenceToProvenancePart01(chain, evidence)
	if binding.Status != WorkloadIdentityProvenanceBindingUnknown || binding.Reason != "PROVENANCE_CHAIN_INCOMPLETE" {
		t.Fatalf("incomplete chain was promoted: %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBindWorkloadIdentityEvidenceRejectsTamperedEvidenceDigest(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01(
		"sha256:"+strings.Repeat("1", 64),
		"sha256:"+strings.Repeat("2", 64),
		"sha256:"+strings.Repeat("3", 64),
		"sha256:"+strings.Repeat("4", 64),
		"sha256:"+strings.Repeat("5", 64),
		"sha256:"+strings.Repeat("6", 64),
	)
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := ObserveWorkloadIdentityEvidence(
		observation,
		"sha256:"+strings.Repeat("b", 64),
		"sha256:"+strings.Repeat("c", 64),
		time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC),
	)
	evidence.EvidenceDigest = "sha256:" + strings.Repeat("d", 64)
	binding := BindWorkloadIdentityEvidenceToProvenancePart01(chain, evidence)
	if binding.Status != WorkloadIdentityProvenanceBindingUnknown || binding.Reason != "IDENTITY_EVIDENCE_DIGEST_INVALID" {
		t.Fatalf("tampered evidence was promoted: %#v", binding)
	}
}
