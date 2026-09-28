package domaincapability

import "testing"

func TestEvidenceEnvelopePreservesPolicyAndBoundary(t *testing.T) {
	policy := Policy{MinimumObservationCount: 2, MaximumUnresolvedRatio: 0.25, RequireEvidenceDigest: true}
	measurement := Measurement{ExpectedCapabilityCount: 4, ObservedCapabilityCount: 3, UnresolvedCapabilityCount: 1, ObservationCount: 3, EvidenceBound: true}
	envelope := NewEvidenceEnvelope(policy, measurement, ReviewMissingCapabilities, "sha256:source", "sha256:contract")
	if envelope.Decision != ReviewMissingCapabilities || !envelope.NonExecuting || !envelope.NonAuthorizing {
		t.Fatalf("envelope lost policy boundary: %+v", envelope)
	}
	first, err := EvidenceEnvelopeDigest(envelope)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EvidenceEnvelopeDigest(envelope)
	if err != nil || first != second {
		t.Fatalf("digest is not deterministic: %q != %q, err=%v", first, second, err)
	}
}