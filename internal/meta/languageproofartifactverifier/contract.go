package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Version: 1, Cases: []CaseSpec{
		{ID: "valid-proof-carrying-artifact", InputKind: "VALID", ExpectedDecision: "PASS", ExpectedResolution: "EXACT", ExpectedReason: "PROOF_CARRYING_ARTIFACT_AUTHORIZED", ProofChoice: "COHERENCE", MetaOperation: "grant-only-after-proof"},
		{ID: "tampered-evidence", InputKind: "TAMPERED", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "PROOF_EVIDENCE_DIGEST_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-tampered-evidence"},
		{ID: "missing-operation-evidence", InputKind: "MISSING", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "PROOF_EVIDENCE_MISSING", ProofChoice: "FOUNDATION", MetaOperation: "lower-missing-evidence"},
		{ID: "bytes-only-no-authority", InputKind: "BYTE_ONLY", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "ARTIFACT_BYTES_NOT_AUTHORITY", ProofChoice: "REGRESSION", MetaOperation: "deny-byte-only-authority"},
		{ID: "independent-recipe-mismatch", InputKind: "WRONG_RECIPE", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "INVARIANT_ONLY", ExpectedReason: "INDEPENDENT_RECIPE_MISMATCH", ProofChoice: "REGRESSION", MetaOperation: "reject-recipe-drift"},
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

func CanonicalRecipe() Recipe {
	return Recipe{Schema: RecipeSchema, Version: 1,
		ID: "gooo://recipe/language-proof-carrying-artifact/v1", Consumer: ConsumerID,
		Steps: []RecipeStep{
			{ID: "verify-source", Input: "source-bytes", MetaOperation: "recheck-source-digest", ProofChoice: "FOUNDATION"},
			{ID: "verify-operation", Input: "operation-receipt", MetaOperation: "recheck-operation-receipt", ProofChoice: "COHERENCE"},
			{ID: "verify-invariant", Input: "invariant-evidence", MetaOperation: "recheck-no-byte-authority", ProofChoice: "REGRESSION"},
			{ID: "grant-authority", Input: "consumer-verdict", MetaOperation: "grant-only-after-proof", ProofChoice: "COHERENCE"},
		},
	}
}
