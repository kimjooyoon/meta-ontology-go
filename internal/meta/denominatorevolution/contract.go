package denominatorevolution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, Version: 1,
		Producer:    "denominatorevolution.Evaluate",
		Consumer:    "denominatorevolutionverify.Verify",
		Denominator: DenominatorSpec{Version: DenominatorVersion, Obligations: canonicalObligations()},
		Policy: MeasurementPolicy{
			NoAggregateEstimates:   true,
			ForbiddenClaims:        []string{"improvement rate", "aggregate estimate", "projected coverage"},
			AllowedAdditionReasons: []string{"NEW_MEASURABLE_OBLIGATION", "GOVERNANCE_GAP_CLOSED"},
			AllowedDeletionReasons: []string{"DEPRECATED_DUPLICATE", "SCOPE_REMOVED", "NOT_MEASURABLE"},
		},
		Cases: []CaseSpec{
			{ID: "legal-advance", Kind: "AUTHORIZED_MIGRATION", ExpectedDecision: "ADVANCE", ExpectedResolution: "EXACT", ExpectedReason: "DENOMINATOR_ADVANCE_AUTHORIZED", FromClaim: "PROPOSED", ToClaim: "ACCEPTED", ProofChoice: "COHERENCE", MetaOperation: "accept-authorized-denominator-advance", Stage: "DECIDE", Step: "apply-migration-receipt", Reason: "both versioned digests and every add-delete reason are bound"},
			{ID: "unauthorized-change", Kind: "UNAUTHORIZED_MUTATION", ExpectedDecision: "BLOCK", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "MIGRATION_RECEIPT_MISSING", FromClaim: "PROPOSED", ToClaim: "REJECTED", ProofChoice: "REGRESSION", MetaOperation: "reject-unreceipted-denominator-change", Stage: "DECIDE", Step: "reject-missing-receipt", Reason: "a changed measurement basis cannot self-authorize"},
			{ID: "unknown-predecessor", Kind: "UNKNOWN_PREDECESSOR", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "UNKNOWN", ExpectedReason: "PREDECESSOR_DIGEST_UNKNOWN", FromClaim: "PROPOSED", ToClaim: "UNKNOWN", ProofChoice: "FOUNDATION", MetaOperation: "fail-closed-unknown-predecessor", Stage: "RESOLVE", Step: "lookup-predecessor", Reason: "an unregistered predecessor is not a legal base"},
		},
		NotClaimed: []string{"improvement rate", "aggregate estimate", "projected coverage", "semantic quality", "repository mutation"},
	}
}

func DecodeContract(raw []byte) (Contract, error) {
	var contract Contract
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("DENOMINATOR_EVOLUTION_CONTRACT_DRIFT")
	}
	return contract, nil
}

func canonicalObligations() []Obligation {
	return []Obligation{
		{ID: "governance/versioned-members", Claim: "The measurement basis has a stable version and exact member set.", Class: "OUTCOME", ProofChoice: "FOUNDATION", MetaOperation: "bind-fixed-denominator", Stage: "FOUNDATION", Step: "pin-versioned-member-set", Reason: "a version and member digest identify exactly what is measured"},
		{ID: "governance/change-reasons", Claim: "Every addition and deletion has an explicit admissible reason.", Class: "DRIVER", ProofChoice: "COHERENCE", MetaOperation: "classify-change-reason", Stage: "PROPOSE", Step: "record-add-delete-reason", Reason: "change intent is evidence, not an inferred improvement"},
		{ID: "governance/predecessor-digest", Claim: "The predecessor version and digest resolve to a known denominator.", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "bind-predecessor-digest", Stage: "MIGRATE", Step: "verify-predecessor-digest", Reason: "a version label without its exact predecessor digest is unknown"},
		{ID: "governance/migration-receipt", Claim: "A migration receipt binds the predecessor, successor, changes, and decision.", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: "accept-migration-receipt", Stage: "MIGRATE", Step: "bind-receipt-to-both-versions", Reason: "the receipt makes a legal denominator transition replayable"},
		{ID: "governance/legacy-summary", Claim: "A human summary may not replace exact denominator members.", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: "reject-summary-substitution", Stage: "GUARD", Step: "preserve-exact-members", Reason: "prose and aggregate summaries cannot authorize a basis change"},
	}
}
