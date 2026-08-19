package main

import (
	"testing"
	"time"
)

func TestGuardianEvidenceRejectsUnavailableOrNonPromotionTopology(t *testing.T) {
	evidence, bundle := validGuardianEvidenceFixture()
	evidence.Topology.Status = "identical"
	if err := validateGuardianEvidence(&evidence, bundle); err == nil {
		t.Fatal("identical main/dev topology was accepted for promotion")
	}
}
func TestGuardianEvidenceRejectsMissingFutureAndExpiredObserverFreshness(t *testing.T) {
	now := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	setFreshness := func(evidence *guardianEvidence, observedAt, validUntil *string) {
		evidence.BranchProtection.ObservedAt = observedAt
		evidence.BranchProtection.ValidUntil = validUntil
		evidence.BranchProtection.Digest = digestBranchProtection(evidence.BranchProtection)
		evidence.DevBranchProtection.ObservedAt = observedAt
		evidence.DevBranchProtection.ValidUntil = validUntil
		evidence.DevBranchProtection.Digest = digestBranchProtection(evidence.DevBranchProtection)
		evidence.ObserverEnvironmentSnapshot.ObservedAt = observedAt
		evidence.ObserverEnvironmentSnapshot.ValidUntil = validUntil
		evidence.ObserverEnvironmentSnapshot.Digest = digestGuardianEnvironment(evidence.ObserverEnvironmentSnapshot)
		evidence.ObserverEnvironmentDigest = evidence.ObserverEnvironmentSnapshot.Digest
		evidence.InstallationRepositoryScope.ObservedAt = observedAt
		evidence.InstallationRepositoryScope.ValidUntil = validUntil
		evidence.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(evidence.InstallationRepositoryScope)
	}
	for name, window := range map[string][2]*string{
		"missing": {nil, stringPointer("2026-08-14T00:10:00Z")},
		"future":  {stringPointer("2026-08-14T00:01:00Z"), stringPointer("2026-08-14T00:11:00Z")},
		"expired": {stringPointer("2026-08-13T23:40:00Z"), stringPointer("2026-08-13T23:50:00Z")},
	} {
		t.Run(name, func(t *testing.T) {
			evidence, bundle := validGuardianEvidenceFixture()
			setFreshness(&evidence, window[0], window[1])
			bundle.BranchProtection = evidence.BranchProtection
			bundle.DevBranchProtection = evidence.DevBranchProtection
			if err := validateGuardianEvidenceAt(&evidence, bundle, now); err == nil {
				t.Fatalf("%s observer freshness was accepted", name)
			}
		})
	}
	evidence, bundle := validGuardianEvidenceFixture()
	setFreshness(&evidence, stringPointer("2026-08-13T23:55:00Z"), stringPointer("2026-08-14T00:05:00Z"))
	bundle.BranchProtection = evidence.BranchProtection
	bundle.DevBranchProtection = evidence.DevBranchProtection
	if err := validateGuardianEvidenceAt(&evidence, bundle, now); err != nil {
		t.Fatalf("valid observer freshness was rejected: %v", err)
	}
}
