package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Version: 2, Cases: []CaseSpec{
		{ID: "valid-proof-carrying-artifact", InputKind: "VALID", ExpectedDecision: "PASS", ExpectedResolution: "EXACT", ExpectedReason: "PROOF_CARRYING_ARTIFACT_AUTHORIZED", ProofChoice: "COHERENCE", MetaOperation: "grant-read-only-consumption"},
		{ID: "tampered-evidence", InputKind: "TAMPERED", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_EVIDENCE_DIGEST_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-tampered-evidence"},
		{ID: "coherent-tamper-reconstruction", InputKind: "COHERENT_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "OPERATION_RECONSTRUCTION_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-coherent-tamper"},
		{ID: "missing-operation-evidence", InputKind: "MISSING", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "PROOF_EVIDENCE_MISSING", ProofChoice: "FOUNDATION", MetaOperation: "lower-missing-evidence"},
		{ID: "bytes-only-no-authority", InputKind: "BYTE_ONLY", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "ARTIFACT_BYTES_NOT_AUTHORITY", ProofChoice: "REGRESSION", MetaOperation: "deny-byte-only-authority"},
		{ID: "independent-recipe-mismatch", InputKind: "WRONG_RECIPE", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "INDEPENDENT_RECIPE_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-recipe-drift"},
		{ID: "recipe-only-mismatch", InputKind: "RECIPE_ONLY", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "INDEPENDENT_RECIPE_MISMATCH", ProofChoice: "COHERENCE", MetaOperation: "reject-recipe-only-drift"},
		{ID: "missing-attachment", InputKind: "MISSING_ATTACHMENT", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "ARTIFACT_ATTACHMENT_MISSING", ProofChoice: "FOUNDATION", MetaOperation: "lower-missing-attachment"},
		{ID: "wrong-attachment-digest", InputKind: "WRONG_ATTACHMENT_DIGEST", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "OPERATION_ATTACHMENT_DIGEST_MISMATCH", ProofChoice: "COHERENCE", MetaOperation: "reject-wrong-attachment"},
		{ID: "unrelated-evidence-tamper", InputKind: "UNRELATED_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "INVARIANT_EVIDENCE_NOT_PRESERVED", ProofChoice: "REGRESSION", MetaOperation: "reject-unrelated-evidence-tamper"},
		{ID: "stale-head", InputKind: "STALE_HEAD", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "HEAD_BINDING_MISMATCH", ProofChoice: "FOUNDATION", MetaOperation: "reject-stale-head"},
		{ID: "unauthorized-consumer", InputKind: "UNAUTHORIZED_CONSUMER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "UNAUTHORIZED_CONSUMER_NOT_ATTESTED", ProofChoice: "REGRESSION", MetaOperation: "deny-unauthorized-consumer"},
		{ID: "coherent-claim-proposition-tamper", InputKind: "CLAIM_PROPOSITION_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_CLAIM_STATEMENT_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-coherent-claim-proposition-tamper"},
		{ID: "coherent-claim-dependency-tamper", InputKind: "CLAIM_DEPENDENCY_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_CLAIM_STATEMENT_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-coherent-claim-dependency-tamper"},
		{ID: "coherent-claim-proof-choice-tamper", InputKind: "CLAIM_PROOF_CHOICE_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_CLAIM_STATEMENT_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-coherent-claim-proof-choice-tamper"},
		{ID: "coherent-claim-target-tamper", InputKind: "CLAIM_TARGET_TAMPER", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_CLAIM_STATEMENT_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-coherent-claim-target-tamper"},
	}}
}

func DecodeContract(raw []byte) (Contract, error) {
	var contract Contract
	if err := json.Unmarshal(raw, &contract); err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("PROOF_CARRYING_ARTIFACT_CONTRACT_DRIFT")
	}
	return contract, nil
}

func CaseIDs() []string {
	return []string{
		"valid-proof-carrying-artifact", "tampered-evidence", "coherent-tamper-reconstruction", "missing-operation-evidence",
		"bytes-only-no-authority", "independent-recipe-mismatch", "recipe-only-mismatch", "missing-attachment",
		"wrong-attachment-digest", "unrelated-evidence-tamper", "stale-head", "unauthorized-consumer",
		"coherent-claim-proposition-tamper", "coherent-claim-dependency-tamper", "coherent-claim-proof-choice-tamper", "coherent-claim-target-tamper",
	}
}

func caseInventoryOK(cases []CaseResult) bool {
	want := CaseIDs()
	if len(cases) != len(want) {
		return false
	}
	for index, item := range cases {
		if item.ID != want[index] {
			return false
		}
	}
	return true
}

func negativeCaseInventoryOK(cases []CaseResult) bool {
	want := CaseIDs()[1:]
	if len(cases) != CaseTotal {
		return false
	}
	for index, item := range cases[1:] {
		if item.ID != want[index] {
			return false
		}
	}
	return true
}

func CanonicalRecipe() Recipe {
	return Recipe{Schema: RecipeSchema, Version: 2,
		ID: "gooo://recipe/language-proof-carrying-artifact/v2", Consumer: ConsumerID, SourceEntry: "GenerateProofCarryingArtifact",
		Roles: []RecipeRole{
			{ID: "source-bytes-bound", Proposition: "source-bytes-match", Target: "raw-source-digest", ProofChoice: "FOUNDATION", Step: "verify-source", MetaOperation: "bind-source-bytes", Dependencies: []string{}},
			{ID: "operation-receipt-bound", Proposition: "operation-receipt-match", Target: "operation-receipt-digest", ProofChoice: "COHERENCE", Step: "verify-operation", MetaOperation: "bind-operation-receipt", Dependencies: []string{"source-bytes-bound"}},
			{ID: "no-byte-authority", Proposition: "generated-bytes-do-not-grant-authority", Target: "read-only-capability-boundary", ProofChoice: "REGRESSION", Step: "verify-invariant", MetaOperation: "preserve-no-byte-authority", Dependencies: []string{}},
			{ID: "recipe-match", Proposition: "consumer-recipe-matches-source-recipe", Target: "recipe-digest", ProofChoice: "COHERENCE", Step: "verify-recipe", MetaOperation: "match-independent-recipe", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority"}},
			{ID: "consumer-authority", Proposition: "verified-consumer-may-read-only-consume", Target: "READ_ONLY_CONSUMPTION", ProofChoice: "COHERENCE", Step: "grant-read-only-consumption", MetaOperation: "grant-read-only-consumption", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match"}},
		},
		Steps: []RecipeStep{
			{ID: "verify-source", Input: "raw-source-digest", MetaOperation: "bind-source-bytes", ProofChoice: "FOUNDATION", Role: "source-bytes-bound"},
			{ID: "verify-operation", Input: "operation-receipt-digest", MetaOperation: "bind-operation-receipt", ProofChoice: "COHERENCE", Role: "operation-receipt-bound"},
			{ID: "verify-invariant", Input: "read-only-capability-boundary", MetaOperation: "preserve-no-byte-authority", ProofChoice: "REGRESSION", Role: "no-byte-authority"},
			{ID: "verify-recipe", Input: "recipe-digest", MetaOperation: "match-independent-recipe", ProofChoice: "COHERENCE", Role: "recipe-match"},
			{ID: "grant-read-only-consumption", Input: "READ_ONLY_CONSUMPTION", MetaOperation: "grant-read-only-consumption", ProofChoice: "COHERENCE", Role: "consumer-authority"},
		},
		Dependencies: []RecipeDependency{
			{From: "source-bytes-bound", To: "operation-receipt-bound", Relation: "requires"},
			{From: "operation-receipt-bound", To: "no-byte-authority", Relation: "requires"},
			{From: "source-bytes-bound,operation-receipt-bound,no-byte-authority", To: "recipe-match", Relation: "requires"},
			{From: "source-bytes-bound,operation-receipt-bound,no-byte-authority,recipe-match", To: "consumer-authority", Relation: "requires"},
		},
		Authority: RecipeAuthority{Capability: "READ_ONLY_CONSUMPTION", Requires: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match"}},
	}
}
