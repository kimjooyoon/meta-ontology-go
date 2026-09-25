package selfimprovementtransport

import "fmt"

func ValidateReport(report Report) error {
	if report.Schema != ReportSchema || report.MetricID != MetricID ||
		report.Contract.ContractID != ContractID || report.Contract.ObligationTotal != fixedObligationTotal ||
		len(report.Obligations) != fixedObligationTotal || report.Metrics.FixedObligationTotal != fixedObligationTotal {
		return fmt.Errorf("transport report shape mismatch")
	}
	if report.Contract.SemanticDigest == "" || report.ProvenanceState != report.Provenance.State ||
		(report.Provenance.State != ResolutionClosed && report.Provenance.State != ResolutionUnknown && report.Provenance.State != ResolutionRefuted) {
		return fmt.Errorf("transport provenance resolution state mismatch")
	}
	if report.Provenance.Stage == "" || report.Provenance.Step == "" || report.Provenance.Reason == "" {
		return fmt.Errorf("transport provenance coordinate is incomplete")
	}
	if err := report.Contract.ResolutionPolicy.Validate(); err != nil {
		return err
	}
	if report.Provenance.State == ResolutionClosed && (report.Provenance.Unknown != nil || report.Provenance.ProducerDeclarationDigest == "" || report.Provenance.ProducerSubjectSHA == "" || report.Provenance.ProducerPayloadDigest == "" || report.Provenance.ProducerPayloadBytes <= 0) {
		return fmt.Errorf("transport CLOSED provenance is not exact")
	}
	if report.Provenance.State == ResolutionUnknown && !causalUnknownComplete(report.Provenance.Unknown) {
		return fmt.Errorf("transport UNKNOWN provenance is missing causal fields")
	}
	if report.Provenance.State == ResolutionRefuted && report.Provenance.Unknown != nil {
		return fmt.Errorf("transport REFUTED provenance carries UNKNOWN evidence")
	}
	if report.ResolutionMetrics != resolutionMetrics(report.Contract.ResolutionPolicy, report.Provenance) {
		return fmt.Errorf("transport resolution metrics mismatch")
	}
	verified, unknown, falseTotal := 0, 0, 0
	for index, obligation := range report.Obligations {
		if obligation.ID != obligationOrder[index] || obligation.Coordinate.Stage == "" || obligation.Coordinate.Step == "" {
			return fmt.Errorf("transport obligation order mismatch")
		}
		switch obligation.Status {
		case StatusVerified:
			verified++
			if !validDigest(obligation.EvidenceDigest) {
				return fmt.Errorf("verified transport evidence digest missing")
			}
		case StatusUnknown:
			unknown++
		case StatusFalse:
			falseTotal++
		default:
			return fmt.Errorf("unknown transport obligation status")
		}
	}
	metrics := report.Metrics
	if metrics.VerifiedTotal != verified || metrics.UnknownTotal != unknown || metrics.FalseTotal != falseTotal ||
		metrics.OpenTotal != unknown+falseTotal || metrics.CoverageBasisPoints != verified*10000/fixedObligationTotal ||
		metrics.FalsePromotionCount != 0 || len(report.OpenObligationIDs) != metrics.OpenTotal {
		return fmt.Errorf("transport metric mismatch")
	}
	switch report.Decision {
	case DecisionPass:
		if verified != fixedObligationTotal || unknown != 0 || falseTotal != 0 || report.Resolution != ResolutionExact {
			return fmt.Errorf("invalid exact transport decision")
		}
	case DecisionObserved:
		if falseTotal != 0 || unknown == 0 || report.Resolution != ResolutionLower {
			return fmt.Errorf("invalid lowered transport decision")
		}
	case DecisionFailClosed:
		if (falseTotal == 0 && report.Provenance.State != ResolutionRefuted) || report.Resolution != ResolutionLower {
			return fmt.Errorf("invalid fail-closed transport decision")
		}
	default:
		return fmt.Errorf("unknown transport decision")
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("transport report digest mismatch")
	}
	return nil
}

func CheckReadOnly(report Report) error {
	if err := ValidateReport(report); err != nil {
		return err
	}
	if report.Decision != DecisionObserved || report.Resolution != ResolutionLower || report.Provenance.State != ResolutionClosed ||
		report.Metrics.VerifiedTotal != 7 || report.Metrics.UnknownTotal != 1 || report.Metrics.FalseTotal != 0 ||
		len(report.OpenObligationIDs) != 1 || report.OpenObligationIDs[0] != attestationObligation ||
		report.Reason != ReasonAttestation {
		return fmt.Errorf("transport is not eligible for unsigned read-only continuation")
	}
	return nil
}

func causalUnknownComplete(value *CausalUnknown) bool {
	return value != nil && value.Stage != "" && value.Step != "" && value.Reason != "" &&
		value.UnknownClass != "" && value.NextOperation != "" && len(value.BlockedBy) > 0
}
