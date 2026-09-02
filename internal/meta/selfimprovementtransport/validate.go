package selfimprovementtransport

import "fmt"

func ValidateReport(report Report) error {
	if report.Schema != ReportSchema || report.MetricID != MetricID ||
		report.Contract.ContractID != ContractID || report.Contract.ObligationTotal != fixedObligationTotal ||
		len(report.Obligations) != fixedObligationTotal || report.Metrics.FixedObligationTotal != fixedObligationTotal {
		return fmt.Errorf("transport report shape mismatch")
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
		if falseTotal == 0 || report.Resolution != ResolutionLower {
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
	if report.Decision != DecisionObserved || report.Resolution != ResolutionLower ||
		report.Metrics.VerifiedTotal != 7 || report.Metrics.UnknownTotal != 1 || report.Metrics.FalseTotal != 0 ||
		len(report.OpenObligationIDs) != 1 || report.OpenObligationIDs[0] != attestationObligation ||
		report.Reason != ReasonAttestation {
		return fmt.Errorf("transport is not eligible for unsigned read-only continuation")
	}
	return nil
}
