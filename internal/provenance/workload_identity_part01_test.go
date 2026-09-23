package provenance

import (
	"maps"
	"strings"
	"testing"
	"time"
)

func TestParseWorkloadIdentityAttributesIsExplicitlyNonAuthorizing(t *testing.T) {
	observation, err := ParseWorkloadIdentityAttributes(map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("a", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	})
	if err != nil {
		t.Fatal(err)
	}
	if observation.ID == "" || observation.AttestationDigest == "" || !observation.NonAuthorizing {
		t.Fatalf("incomplete identity observation: %#v", observation)
	}
	if !observation.ExpiresAt.Equal(time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expiry = %s", observation.ExpiresAt)
	}
	if got := observation.FreshnessAt(time.Date(2026, 9, 24, 11, 59, 59, 0, time.UTC)); got != WorkloadIdentityFreshnessFresh {
		t.Fatalf("freshness before expiry = %q", got)
	}
	if got := observation.FreshnessAt(time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)); got != WorkloadIdentityFreshnessExpired {
		t.Fatalf("freshness at expiry = %q", got)
	}
}

func TestParseWorkloadIdentityAttributesRejectsAuthorityAndShapeConfusion(t *testing.T) {
	base := map[string]string{
		"workload_identity":            "spiffe://example.org/ns/billing/sa/worker",
		"workload_attestation_digest":  "sha256:" + strings.Repeat("b", 64),
		"workload_identity_expires_at": "2026-09-24T12:00:00Z",
		"workload_identity_authority":  WorkloadIdentityAuthorityNonAuthorizing,
	}
	for name, mutate := range map[string]func(map[string]string){
		"missing authority": func(value map[string]string) { delete(value, "workload_identity_authority") },
		"wrong scheme":      func(value map[string]string) { value["workload_identity"] = "https://example.org/workload" },
		"invalid digest":    func(value map[string]string) { value["workload_attestation_digest"] = "sha256:not-a-digest" },
	} {
		t.Run(name, func(t *testing.T) {
			value := make(map[string]string, len(base))
			maps.Copy(value, base)
			mutate(value)
			if _, err := ParseWorkloadIdentityAttributes(value); err == nil {
				t.Fatal("invalid workload identity observation was accepted")
			}
		})
	}
}

func TestWorkloadIdentityFreshnessRequiresObservationAndClock(t *testing.T) {
	var observation WorkloadIdentityObservation
	for _, now := range []time.Time{time.Time{}, time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)} {
		if got := observation.FreshnessAt(now); got != WorkloadIdentityFreshnessUnknown {
			t.Fatalf("empty observation freshness = %q", got)
		}
	}
}
