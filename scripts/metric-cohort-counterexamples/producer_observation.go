package main

import "fmt"

func producerObservation(row countercase, v variant, f fixture) bool {
	expected := len(f.Contract)
	if row.Kind != "PRODUCER_INPUT_MUTATION" {
		return !v.ProducerAttempted && v.Producer == nil && v.ProducerError == ""
	}
	if !v.ProducerAttempted {
		return false
	}
	if row.ID == "ambiguous-binding" {
		return v.Producer == nil && len(v.Receipt) == 0 && expectedDuplicateReason(f) != "" &&
			v.ProducerError == expectedDuplicateReason(f)
	}
	r := v.Producer
	if v.ProducerError != "" || r == nil || r.ExpectedCount != expected || r.Status != "FAIL_CLOSED" ||
		r.FailureClass != "UNKNOWN" || r.Unknown == nil || r.RepositoryWorkspaceWrites || r.PromotionAuthorized {
		return false
	}
	switch row.ID {
	case "missing-observed-binding":
		return r.BoundCount == expected-1 && r.UnknownCount == 0 && r.CoverageBPS == (expected-1)*10000/expected &&
			r.Unknown.UnknownClass == "DIRECT_MISSING" && r.Unknown.Reason == "CONCEPT_OPERATION_DIRECT_EVIDENCE_MISSING"
	case "replaced-observed-binding":
		return r.BoundCount == expected-1 && r.UnknownCount == 1 && r.CoverageBPS == (expected-1)*10000/expected &&
			r.Unknown.UnknownClass == "AMBIGUOUS" && r.Unknown.Reason == "CONCEPT_OPERATION_BINDING_AMBIGUOUS"
	case "extra-unknown-observation":
		return r.BoundCount == expected && r.UnknownCount == 1 && r.CoverageBPS == 10000 &&
			r.Unknown.UnknownClass == "AMBIGUOUS" && r.Unknown.NextOperation == "DISAMBIGUATE_CONCEPT_OPERATION_BINDINGS"
	}
	return false
}

func expectedDuplicateReason(f fixture) string {
	for _, binding := range f.Receipt.Expected {
		if binding.ProofChoice == "FOUNDATION" {
			return fmt.Sprintf("duplicate indicator %q", binding.IndicatorID)
		}
	}
	return ""
}
