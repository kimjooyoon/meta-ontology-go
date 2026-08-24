package sourceauthority

import "fmt"

func validateMeasurement(contract Contract) error {
	measurement := contract.Measurement
	if measurement.Numerator !=
		"accepted_facts_with_exact_source_span_digest_and_authority_ref" {
		return fmt.Errorf("unexpected numerator %q", measurement.Numerator)
	}
	if measurement.Denominator != "accepted_facts" {
		return fmt.Errorf("unexpected denominator %q", measurement.Denominator)
	}
	if measurement.Unit != "basis_points" || measurement.Target != 10000 {
		return fmt.Errorf("measurement target must be 10000 basis_points")
	}
	if contract.AdoptionState != "CONTRACT_ONLY" || contract.ReadinessCredit != 0 {
		return fmt.Errorf("contract epoch must grant zero readiness credit")
	}
	scope := contract.Scope
	if scope.AcceptedFactMode != "EXACT_SOURCE_BYTES_ONLY" {
		return fmt.Errorf("accepted facts must use exact source bytes")
	}
	if scope.SemanticInterpretation != "CANDIDATE_ONLY" {
		return fmt.Errorf("semantic interpretation must remain candidate")
	}
	if scope.LiveURLWithoutSnapshot != "UNKNOWN" {
		return fmt.Errorf("live URL without snapshot must remain unknown")
	}
	return nil
}
