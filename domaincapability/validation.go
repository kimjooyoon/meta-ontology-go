package domaincapability

import "fmt"

// ValidateEvidenceEnvelope rejects missing provenance and boundary violations.
func ValidateEvidenceEnvelope(envelope EvidenceEnvelope) error {
	if envelope.Schema != EvidenceEnvelopeSchema {
		return fmt.Errorf("schema %q does not match %q", envelope.Schema, EvidenceEnvelopeSchema)
	}
	if envelope.SourceDigest == "" || envelope.ContractDigest == "" {
		return fmt.Errorf("source and contract digests are required")
	}
	if !envelope.NonExecuting || !envelope.NonAuthorizing {
		return fmt.Errorf("evidence envelope cannot authorize or execute")
	}
	if envelope.ExpectedCapabilityCount < 0 || envelope.ObservedCapabilityCount < 0 || envelope.UnresolvedCapabilityCount < 0 || envelope.ObservationCount < 0 {
		return fmt.Errorf("measurement counts cannot be negative")
	}
	if envelope.ObservedCapabilityCount > envelope.ExpectedCapabilityCount {
		return fmt.Errorf("observed capability count exceeds expected boundary")
	}
	if envelope.UnresolvedCapabilityCount > envelope.ExpectedCapabilityCount {
		return fmt.Errorf("unresolved capability count exceeds expected boundary")
	}
	if envelope.MaximumUnresolvedRatio < -1 || envelope.MaximumUnresolvedRatio > 1 {
		return fmt.Errorf("maximum unresolved ratio is outside the supported range")
	}
	switch envelope.Decision {
	case CollectMoreObservations, ReviewMissingCapabilities, ReviewInvestment, NoActionableSignal:
		return nil
	default:
		return fmt.Errorf("unknown decision %q", envelope.Decision)
	}
}