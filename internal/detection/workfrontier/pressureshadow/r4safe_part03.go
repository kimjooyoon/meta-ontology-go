package pressureshadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func finishR4Safe(result R4SafeResult, decision Decision, reason Reason, fullSuite bool) R4SafeResult {
	result.Decision, result.Reason, result.FullSuiteRequired = decision, reason, fullSuite
	result.ExecutionAuthorized, result.EnforcementEffect = false, EnforcementNoEffect
	if decision != DecisionPass {
		result.SafeSelectedIDs, result.SafeWorkIDs = nil, nil
	}
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	result.ResultDigest = digestBytes(data)
	result.ReplayDigest = digestBytes([]byte("r4-safe-replay\x00" + result.InputDigest + "\x00" +
		result.R4ResultDigest + "\x00" + result.PressureResultDigest + "\x00" + result.ResultDigest))
	return result
}
func CanonicalR4SafeInputBytes(input R4SafeInput) ([]byte, error) {
	if err := validateR4SafeSyntax(input); err != nil {
		return nil, err
	}
	r4Bytes, err := workfrontier.EncodeR4JSON(input.R4Input)
	if err != nil {
		return nil, err
	}
	shadow, err := CanonicalInputBytes(projectR4SafeInput(input))
	if err != nil {
		return nil, err
	}
	var rows map[string]json.RawMessage
	if err := json.Unmarshal(shadow, &rows); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Schema       string          `json:"schema"`
		R4Input      json.RawMessage `json:"r4_input"`
		PathCoverage json.RawMessage `json:"path_coverage"`
	}{input.Schema, bytes.TrimSpace(r4Bytes), rows["path_coverage"]})
}
func validateR4SafeSyntax(input R4SafeInput) error {
	if input.Schema != R4SafeSchemaVersion || input.R4Input.SchemaVersion != workfrontier.R4SchemaVersion {
		return fmt.Errorf("invalid schema")
	}
	return validateSyntax(projectR4SafeInput(input))
}
func inspectR4Safe(input R4SafeInput) (invalid, missing bool) {
	projected := projectR4SafeInput(input)
	_, extra, _, _ := b1Issues(projected)
	if len(orphanPathIDs(selectorPathIDs(projected), coverageRows(projected)))+len(extra) > 0 {
		return true, false
	}
	var registered, recorded []string
	for _, pressure := range input.R4Input.Pressures {
		registered = append(registered, pressureID(pressure))
	}
	for _, row := range input.PathCoverage {
		for _, record := range row.Coverage.PressureRecords {
			recorded = append(recorded, record.PressureID)
		}
	}
	if len(pressureDifference(recorded, registered)) > 0 {
		return true, false
	}
	return false, len(pressureDifference(registered, recorded)) > 0
}
