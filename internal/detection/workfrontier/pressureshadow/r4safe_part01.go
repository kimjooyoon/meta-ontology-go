package pressureshadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"io"
)

const R4SafeSchemaVersion = "gooo/workfrontier-pressure-safe-r4/v1"

type R4SafeInput struct {
	Schema       string               `json:"schema"`
	R4Input      workfrontier.R4Input `json:"r4_input"`
	PathCoverage []PathCoverage       `json:"path_coverage"`
}

func (input *R4SafeInput) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoded, err := decodeR4SafeWire(data)
	if err != nil {
		return err
	}
	if err := validateR4SafeSyntax(decoded); err != nil {
		return err
	}
	*input = decoded
	return nil
}
func decodeR4SafeWire(data []byte) (R4SafeInput, error) {
	type wire R4SafeInput
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wire
	if err := decoder.Decode(&decoded); err != nil {
		return R4SafeInput{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return R4SafeInput{}, fmt.Errorf("trailing JSON value")
	}
	return R4SafeInput(decoded), nil
}

type R4SafeResult struct {
	Schema               string                `json:"schema"`
	InputDigest          string                `json:"input_digest"`
	R4Result             workfrontier.R4Result `json:"r4_result"`
	R4ResultDigest       string                `json:"r4_result_digest"`
	PressureResult       S1B1Result            `json:"pressure_result"`
	PressureResultDigest string                `json:"pressure_result_digest"`
	Decision             Decision              `json:"decision"`
	Reason               Reason                `json:"reason"`
	FullSuiteRequired    bool                  `json:"full_suite_required"`
	SafeSelectedIDs      []string              `json:"safe_selected_ids"`
	SafeWorkIDs          []string              `json:"safe_work_ids"`
	ExecutionAuthorized  bool                  `json:"execution_authorized"`
	EnforcementEffect    EnforcementEffect     `json:"enforcement_effect"`
	ResultDigest         string                `json:"result_digest"`
	ReplayDigest         string                `json:"replay_digest"`
}
