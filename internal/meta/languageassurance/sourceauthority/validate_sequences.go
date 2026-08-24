package sourceauthority

import "fmt"

var expectedRules = []string{
	"accepted_fact_requires_source_ref",
	"source_ref_requires_exact_snapshot_digest",
	"accepted_fact_requires_exact_byte_span",
	"span_digest_must_match_source_bytes",
	"accepted_fact_requires_authority_ref",
	"authority_ref_must_authorize_exact_source_scope",
	"semantic_interpretation_remains_candidate",
}

func validateSequences(contract Contract) error {
	checks := []struct {
		name string
		got  []string
		want []string
	}{
		{"observation", contract.States.Observation,
			[]string{"SATISFIED", "NOT_SATISFIED", "UNKNOWN", "ERROR"}},
		{"resolution", contract.States.Resolution,
			[]string{"EXACT", "INVARIANT_ONLY"}},
		{"enforcement", contract.States.Enforcement,
			[]string{"ALLOW", "BLOCK", "NO_EFFECT"}},
		{"rules", contract.Rules, expectedRules},
	}
	for _, check := range checks {
		if !sameStrings(check.got, check.want) {
			return fmt.Errorf("%s sequence does not match v1", check.name)
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validateFailureModes(contract Contract) error {
	if err := requireBlock(contract.UnknownEvidence,
		"SOURCE_AUTHORITY_EVIDENCE_UNKNOWN"); err != nil {
		return err
	}
	return requireBlock(contract.EmptyDenominator,
		"SOURCE_AUTHORITY_DENOMINATOR_EMPTY")
}

func requireBlock(mode FailureMode, reason string) error {
	if mode.Observation != "UNKNOWN" ||
		mode.Resolution != "INVARIANT_ONLY" ||
		mode.Enforcement != "BLOCK" || mode.Reason != reason {
		return fmt.Errorf("%s must preserve UNKNOWN at lower resolution and BLOCK", reason)
	}
	return nil
}
