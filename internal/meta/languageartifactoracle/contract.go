package languageartifactoracle

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const ContractSchema = "gooo/language-artifact-oracle-contract/v1"

type Contract struct {
	Schema  string     `json:"schema"`
	Version int        `json:"version"`
	Cases   []CaseSpec `json:"cases"`
}

type CaseSpec struct {
	ID               string `json:"id"`
	InputKind        string `json:"input_kind"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedReason   string `json:"expected_reason"`
	ProofChoice      string `json:"proof_choice"`
	MetaOperation    string `json:"meta_operation"`
}

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Version: 1, Cases: []CaseSpec{
		{"genuine-source-bound", "GENUINE", "PASS", "ARTIFACT_SOURCE_PROJECTION_EXACT", "FOUNDATION", "independently-project-source"},
		{"resealed-output-forgery", "FORGED", "FAIL_CLOSED", "ARTIFACT_SOURCE_PROJECTION_MISMATCH", "REGRESSION", "reject-resealed-artifact-forgery"},
		{"unknown-artifact-decision", "UNKNOWN_DECISION", "FAIL_CLOSED", "ARTIFACT_DECISION_UNKNOWN", "REGRESSION", "reject-unknown-artifact-decision"},
		{"unsupported-source-projection", "UNSUPPORTED_SOURCE", "FAIL_CLOSED", "ORACLE_SOURCE_PROJECTION_UNKNOWN", "FOUNDATION", "lower-unsupported-source-resolution"},
	}}
}

func DecodeContract(raw []byte) (Contract, error) {
	var contract Contract
	if err := json.Unmarshal(raw, &contract); err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("ARTIFACT_ORACLE_CONTRACT_DRIFT")
	}
	return contract, nil
}
