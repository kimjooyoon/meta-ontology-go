package ciplanusecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func FixedContract() Contract {
	return Contract{
		Schema: ContractSchema, Denominator: 12,
		Cases: []CaseSpec{
			{ID: "pass-combined", ExpectedDecision: "PASS", ProofChoice: "COHERENCE"},
			{ID: "pass-docs", ExpectedDecision: "PASS", ProofChoice: "COHERENCE"},
			{ID: "pass-go", ExpectedDecision: "PASS", ProofChoice: "COHERENCE"},
			{ID: "pass-yaml", ExpectedDecision: "PASS", ProofChoice: "COHERENCE"},
			{ID: "fail-absolute", ExpectedDecision: "FAIL_CLOSED", ProofChoice: "REGRESSION"},
			{ID: "fail-duplicate", ExpectedDecision: "FAIL_CLOSED", ProofChoice: "REGRESSION"},
			{ID: "fail-empty", ExpectedDecision: "FAIL_CLOSED", ProofChoice: "REGRESSION"},
			{ID: "fail-traversal", ExpectedDecision: "FAIL_CLOSED", ProofChoice: "REGRESSION"},
			{ID: "unknown-asset", ExpectedDecision: "UNKNOWN", ProofChoice: "FOUNDATION"},
			{ID: "unknown-license", ExpectedDecision: "UNKNOWN", ProofChoice: "FOUNDATION"},
			{ID: "unknown-python", ExpectedDecision: "UNKNOWN", ProofChoice: "FOUNDATION"},
			{ID: "unknown-toml", ExpectedDecision: "UNKNOWN", ProofChoice: "FOUNDATION"},
		},
		Limits: Limits{MaxWallMS: 5000, MaxPeakRSSKiB: 262144, MaxReceiptBytes: 65536},
		NotClaimed: []string{"check-execution", "full-language-semantic-correctness", "general-build-planning", "comparative-performance", "production-or-external-effects"},
	}
}

func DecodeContract(raw []byte) (Contract, error) {
	contract := Contract{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("decode ci-plan contract: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Contract{}, fmt.Errorf("decode ci-plan contract: trailing content")
	}
	if !reflect.DeepEqual(contract, FixedContract()) {
		return Contract{}, fmt.Errorf("ci-plan contract differs from fixed 12-case denominator")
	}
	return contract, nil
}

func DecodeGolden(raw []byte) (GoldenPlan, error) {
	golden := GoldenPlan{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&golden); err != nil {
		return GoldenPlan{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return GoldenPlan{}, fmt.Errorf("golden plan contains trailing content")
	}
	if golden.Schema != "gooo/ci-plan-golden/v1" || golden.CaseID == "" {
		return GoldenPlan{}, fmt.Errorf("golden plan identity is invalid")
	}
	return golden, nil
}
