package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const WorkloadIdentityEvidenceSchema = "gooo/workload-identity-evidence/v1"

type WorkloadIdentityEvidenceStatus string

const (
	WorkloadIdentityEvidenceUnknown  WorkloadIdentityEvidenceStatus = "UNKNOWN"
	WorkloadIdentityEvidenceObserved WorkloadIdentityEvidenceStatus = "OBSERVED"
	WorkloadIdentityEvidenceExpired  WorkloadIdentityEvidenceStatus = "EXPIRED"
)

// WorkloadIdentityEvidence binds identity, SVID, and trust-bundle digests
// without claiming that cryptographic verification or authorization occurred.
type WorkloadIdentityEvidence struct {
	Schema            string                         `json:"schema"`
	ID                string                         `json:"id"`
	TrustDomain       string                         `json:"trust_domain,omitempty"`
	AttestationDigest string                         `json:"attestation_digest,omitempty"`
	SVIDDigest        string                         `json:"svid_digest,omitempty"`
	TrustBundleDigest string                         `json:"trust_bundle_digest,omitempty"`
	Freshness         WorkloadIdentityFreshness      `json:"freshness"`
	Status            WorkloadIdentityEvidenceStatus `json:"status"`
	Reason            string                         `json:"reason"`
	NonAuthorizing    bool                           `json:"non_authorizing"`
	EvidenceDigest    string                         `json:"evidence_digest"`
}

// ObserveWorkloadIdentityEvidence combines identity observations with the
// external SVID and trust-bundle identities. It never verifies either object.
func ObserveWorkloadIdentityEvidence(observation WorkloadIdentityObservation, svidDigest, trustBundleDigest string, now time.Time) WorkloadIdentityEvidence {
	evidence := WorkloadIdentityEvidence{
		Schema:            WorkloadIdentityEvidenceSchema,
		ID:                strings.TrimSpace(observation.ID),
		AttestationDigest: strings.TrimSpace(observation.AttestationDigest),
		SVIDDigest:        strings.TrimSpace(svidDigest),
		TrustBundleDigest: strings.TrimSpace(trustBundleDigest),
		Freshness:         observation.FreshnessAt(now),
		Status:            WorkloadIdentityEvidenceUnknown,
		Reason:            "IDENTITY_EVIDENCE_INCOMPLETE",
		NonAuthorizing:    true,
	}
	if parsed, err := url.Parse(evidence.ID); err == nil && parsed.Scheme == "spiffe" {
		evidence.TrustDomain = parsed.Host
	}
	switch {
	case !observation.NonAuthorizing || evidence.ID == "" || evidence.TrustDomain == "":
		evidence.Reason = "IDENTITY_OBSERVATION_INVALID"
	case !isSHA256Digest(evidence.AttestationDigest):
		evidence.Reason = "MISSING_ATTESTATION_DIGEST"
	case !isSHA256Digest(evidence.SVIDDigest):
		evidence.Reason = "MISSING_SVID_DIGEST"
	case !isSHA256Digest(evidence.TrustBundleDigest):
		evidence.Reason = "MISSING_TRUST_BUNDLE_DIGEST"
	case evidence.Freshness == WorkloadIdentityFreshnessUnknown:
		evidence.Reason = "IDENTITY_FRESHNESS_UNKNOWN"
	case evidence.Freshness == WorkloadIdentityFreshnessExpired:
		evidence.Status = WorkloadIdentityEvidenceExpired
		evidence.Reason = "IDENTITY_SVID_EXPIRED"
	default:
		evidence.Status = WorkloadIdentityEvidenceObserved
		evidence.Reason = "IDENTITY_AND_BUNDLE_IDENTITIES_OBSERVED"
	}
	evidence.EvidenceDigest = workloadIdentityEvidenceDigest(evidence)
	return evidence
}

func workloadIdentityEvidenceDigest(value WorkloadIdentityEvidence) string {
	value.EvidenceDigest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
