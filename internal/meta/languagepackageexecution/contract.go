package languagepackageexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Version: 1, Cases: []CaseSpec{
		{ID: "positive-package-execution", ExpectedDecision: "PASS", ExpectedReason: "PACKAGE_EXECUTED", ProofChoice: "FOUNDATION"},
		{ID: "deterministic-replay", ExpectedDecision: "PASS", ExpectedReason: "PACKAGE_EXECUTED", ProofChoice: "REGRESSION"},
		{ID: "header-mismatch-rejection", ExpectedDecision: "FAIL_CLOSED", ExpectedReason: "PACKAGE_HEADER_MISMATCH", ProofChoice: "COHERENCE"},
		{ID: "duplicate-declaration-rejection", ExpectedDecision: "FAIL_CLOSED", ExpectedReason: "PACKAGE_EXECUTION_REJECTED", ProofChoice: "COHERENCE"},
		{ID: "source-count-rejection", ExpectedDecision: "FAIL_CLOSED", ExpectedReason: "PACKAGE_SOURCE_COUNT_INVALID", ProofChoice: "FOUNDATION"},
	}}
}

func DecodeContract(data []byte) (Contract, error) {
	var contract Contract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, err
	}
	if !reflect.DeepEqual(contract, CanonicalContract()) {
		return Contract{}, fmt.Errorf("languagepackageexecution: contract differs from canonical fixed denominator")
	}
	return contract, nil
}
