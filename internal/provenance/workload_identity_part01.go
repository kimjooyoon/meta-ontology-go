package provenance

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const WorkloadIdentityAuthorityNonAuthorizing = "non_authorizing"

type WorkloadIdentityFreshness string

const (
	WorkloadIdentityFreshnessUnknown WorkloadIdentityFreshness = "UNKNOWN"
	WorkloadIdentityFreshnessFresh   WorkloadIdentityFreshness = "FRESH"
	WorkloadIdentityFreshnessExpired WorkloadIdentityFreshness = "EXPIRED"
)

type WorkloadIdentityObservation struct {
	ID                string
	AttestationDigest string
	ExpiresAt         time.Time
	NonAuthorizing    bool
}

// ParseWorkloadIdentityAttributes validates an optional workload identity
// observation without treating it as authentication or CI authority. The
// cryptographic SVID and trust-bundle verification remain external concerns.
func ParseWorkloadIdentityAttributes(attributes map[string]string) (WorkloadIdentityObservation, error) {
	if len(attributes) == 0 {
		return WorkloadIdentityObservation{}, nil
	}
	identity := strings.TrimSpace(attributes["workload_identity"])
	attestation := strings.TrimSpace(attributes["workload_attestation_digest"])
	expires := strings.TrimSpace(attributes["workload_identity_expires_at"])
	authority := strings.TrimSpace(attributes["workload_identity_authority"])
	if identity == "" && attestation == "" && expires == "" && authority == "" {
		return WorkloadIdentityObservation{}, nil
	}
	if identity == "" || attestation == "" || expires == "" {
		return WorkloadIdentityObservation{}, fmt.Errorf("workload identity observation is incomplete")
	}
	if authority != WorkloadIdentityAuthorityNonAuthorizing {
		return WorkloadIdentityObservation{}, fmt.Errorf("workload identity authority must be %q", WorkloadIdentityAuthorityNonAuthorizing)
	}
	parsed, err := url.Parse(identity)
	if err != nil || parsed.Scheme != "spiffe" || parsed.Host == "" || parsed.Path == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return WorkloadIdentityObservation{}, fmt.Errorf("workload identity must be a path-bearing spiffe URI")
	}
	if !isSHA256Digest(attestation) {
		return WorkloadIdentityObservation{}, fmt.Errorf("workload attestation digest must be sha256")
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, expires)
	if err != nil {
		return WorkloadIdentityObservation{}, fmt.Errorf("workload identity expiry is invalid: %w", err)
	}
	return WorkloadIdentityObservation{
		ID: identity, AttestationDigest: attestation, ExpiresAt: expiresAt,
		NonAuthorizing: true,
	}, nil
}

func (observation WorkloadIdentityObservation) FreshnessAt(now time.Time) WorkloadIdentityFreshness {
	if observation.ID == "" || observation.ExpiresAt.IsZero() || now.IsZero() {
		return WorkloadIdentityFreshnessUnknown
	}
	if now.Before(observation.ExpiresAt) {
		return WorkloadIdentityFreshnessFresh
	}
	return WorkloadIdentityFreshnessExpired
}

func isSHA256Digest(value string) bool {
	algorithm, encoded, ok := strings.Cut(value, ":")
	if !ok || algorithm != "sha256" || len(encoded) != 64 {
		return false
	}
	_, err := hex.DecodeString(encoded)
	return err == nil
}
