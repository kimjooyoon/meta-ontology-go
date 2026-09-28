package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func measurementDigest(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestMeasureDomainBindsScopeWithoutClaimingCompleteness(t *testing.T) {
	plan := MeasureDomain(DomainMeasurementInput{
		Domain:                    "language-operations",
		ScopeDigest:               measurementDigest("scope"),
		SourceDigest:               measurementDigest("source"),
		CapabilityDigest:           measurementDigest("capabilities"),
		ObservationEvidenceDigest: measurementDigest("observations"),
		ObservedUseCases:           2,
		RequiredUseCases:           5,
		InvestmentBudget:           8,
		EstimatedInvestment:        3,
	})
	if plan.Status != DomainMeasurementBound || plan.Decision != DomainDecisionAddUseCase {
		t.Fatalf("unexpected bound measurement: %+v", plan)
	}
	if plan.RemainingUseCases != 3 || plan.MetricMeaning == "language completeness" {
		t.Fatalf("measurement should preserve a scoped evidence gap: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("bound measurement should validate: %v", err)
	}
}

func TestMeasureDomainDefersInvestmentBeyondBudget(t *testing.T) {
	plan := MeasureDomain(DomainMeasurementInput{
		Domain:                    "language-operations",
		ScopeDigest:               measurementDigest("scope"),
		SourceDigest:               measurementDigest("source"),
		CapabilityDigest:           measurementDigest("capabilities"),
		ObservationEvidenceDigest: measurementDigest("observations"),
		ObservedUseCases:           5,
		RequiredUseCases:           5,
		InvestmentBudget:           2,
		EstimatedInvestment:        3,
	})
	if plan.Status != DomainMeasurementBound || plan.Decision != DomainDecisionDeferInvestment {
		t.Fatalf("unexpected budget decision: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("budget decision should validate: %v", err)
	}
}

func TestMeasureDomainPreservesMissingScopeBoundary(t *testing.T) {
	plan := MeasureDomain(DomainMeasurementInput{
		Domain:                    "language-operations",
		SourceDigest:               measurementDigest("source"),
		CapabilityDigest:           measurementDigest("capabilities"),
		ObservationEvidenceDigest: measurementDigest("observations"),
		ObservedUseCases:           1,
		RequiredUseCases:           2,
		InvestmentBudget:           1,
		EstimatedInvestment:        1,
	})
	if plan.Status != DomainMeasurementDeferred || plan.FirstBoundary != "scope" {
		t.Fatalf("unexpected deferred measurement: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("deferred measurement should validate: %v", err)
	}
}
