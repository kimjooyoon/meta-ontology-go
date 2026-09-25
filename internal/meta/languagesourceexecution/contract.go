package languagesourceexecution

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const ContractSchema = "gooo/language-source-execution-contract/v1"

type Contract struct {
	Schema  string     `json:"schema"`
	Version int        `json:"version"`
	Cases   []CaseSpec `json:"cases"`
}

type CaseSpec struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedReason   string `json:"expected_reason"`
	ProofChoice      string `json:"proof_choice"`
	MetaOperation    string `json:"meta_operation"`
}

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Version: 1, Cases: []CaseSpec{
		{"execute-billing", "POSITIVE", "PASS", "SOURCE_ACTIVITY_EXECUTED", "FOUNDATION", "execute-source-activity"},
		{"deterministic-replay", "POSITIVE", "PASS", "SOURCE_ACTIVITY_EXECUTED", "COHERENCE", "replay-source-execution-result"},
		{"unknown-entry", "GUARDRAIL", "FAIL_CLOSED", "SOURCE_ENTRY_UNKNOWN", "REGRESSION", "reject-source-runtime-failure"},
		{"invalid-syntax", "GUARDRAIL", "FAIL_CLOSED", "SOURCE_SYNTAX_INVALID", "REGRESSION", "reject-source-runtime-failure"},
	}}
}

func DecodeContract(raw []byte) (Contract, error) {
	var contract Contract
	if err := json.Unmarshal(raw, &contract); err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("SOURCE_EXECUTION_CONTRACT_DRIFT")
	}
	return contract, nil
}
