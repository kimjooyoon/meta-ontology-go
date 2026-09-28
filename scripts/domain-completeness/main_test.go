package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func validDomainReceipt() domainReceipt {
	digest := strings.Repeat("a", 64)
	return domainReceipt{
		DomainID: "gooo-capabilities",
		DomainVersion: "v1",
		CorrelationID: "corr-1",
		ObservedAt: "2026-09-28T06:00:00Z",
		DeclarationDigest: digest,
		SourceDigest: digest,
		CatalogDigest: digest,
		SchemaDigest: digest,
		EvidenceDigest: digest,
		ToolchainIdentity: "gooo-jev/v1",
		DeclaredIdentifiers: []string{"ir", "provenance"},
		Observations: []domainObservation{
			{ID: "ir", State: "AVAILABLE"},
			{ID: "provenance", State: "UNKNOWN", Reason: "reverse observation not supplied"},
		},
		Metrics: domainMetrics{
			ScopeCoverage: boundedMetric{Value: 1, Numerator: 2, Denominator: 2},
			AvailableCoverage: boundedMetric{Value: 0.5, Numerator: 1, Denominator: 2},
			EvidenceCompleteness: boundedMetric{Value: 1, Numerator: 2, Denominator: 2},
			UtilityYield: boundedMetric{Value: 0, Numerator: 0, Denominator: 0},
		},
		Provenance: domainProvenance{FirstMissingStage: 2, NextOperation: "supply_reverse_observation"},
	}
}

func TestDomainReceiptValidatesBoundedMetricsAndUnknown(t *testing.T) {
	receipt := validDomainReceipt()
	if err := receipt.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := receipt.Digest(); err != nil {
		t.Fatalf("Digest() error = %v", err)
	}
}

func TestDomainReceiptRejectsCompletionScoreDrift(t *testing.T) {
	receipt := validDomainReceipt()
	receipt.Metrics.AvailableCoverage.Value = 1
	if err := receipt.Validate(); err == nil {
		t.Fatal("Validate() accepted a metric value unrelated to its numerator")
	}
}

func TestDomainReceiptRejectsUnknownWithoutCause(t *testing.T) {
	receipt := validDomainReceipt()
	receipt.Observations[1].Reason = ""
	if err := receipt.Validate(); err == nil {
		t.Fatal("Validate() accepted UNKNOWN without a cause")
	}
}

func TestDecodeReceiptRejectsTrailingJSON(t *testing.T) {
	payload, err := json.Marshal(validDomainReceipt())
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	payload = append(payload, []byte(`{}`)...)
	if _, err := decodeReceipt(payload); err == nil {
		t.Fatal("decodeReceipt() accepted trailing JSON")
	}
}