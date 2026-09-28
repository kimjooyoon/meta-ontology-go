package domaincapability

import "testing"

func TestValidateEvidenceEnvelopeRejectsBoundaryCrossing(t *testing.T) {
	envelope := NewEvidenceEnvelope(
		Policy{MinimumObservationCount: 1, MaximumUnresolvedRatio: 0.5, RequireEvidenceDigest: true},
		Measurement{ExpectedCapabilityCount: 1, ObservedCapabilityCount: 1, ObservationCount: 1, EvidenceBound: true},
		ReviewInvestment,
		"sha256:source",
		"sha256:contract",
	)
	envelope.NonAuthorizing = false
	if err := ValidateEvidenceEnvelope(envelope); err == nil {
		t.Fatal("expected authorization boundary rejection")
	}
}

func TestValidateEvidenceEnvelopeAcceptsBoundedMeasurement(t *testing.T) {
	envelope := NewEvidenceEnvelope(
		Policy{MinimumObservationCount: 1, MaximumUnresolvedRatio: 0.5, RequireEvidenceDigest: true},
		Measurement{ExpectedCapabilityCount: 2, ObservedCapabilityCount: 1, UnresolvedCapabilityCount: 1, ObservationCount: 1, EvidenceBound: true},
		ReviewMissingCapabilities,
		"sha256:source",
		"sha256:contract",
	)
	if err := ValidateEvidenceEnvelope(envelope); err != nil {
		t.Fatal(err)
	}
}