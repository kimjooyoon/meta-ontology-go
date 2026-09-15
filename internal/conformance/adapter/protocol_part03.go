package adapter

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c Contract) validate() error {
	fields := []struct {
		name  string
		value string
	}{
		{"ast contract", c.AST}, {"ir contract", c.IR},
		{"generator contract", c.Generator}, {"marker contract", c.Marker},
		{"policy digest", c.PolicyHash},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	return nil
}
func (e Expectation) validate() error {
	if !validStatus(e.Status) {
		return fmt.Errorf("unsupported expected status %q", e.Status)
	}
	if e.Status != StatusFail && e.FailureCode != "" {
		return fmt.Errorf("failure code requires expected fail status")
	}
	return nil
}
func validateAuthoritativeInput(operation Operation, input Input) error {
	if operation == OperationParseAST && strings.TrimSpace(input.DSL) == "" {
		return fmt.Errorf("parse-ast requires DSL input")
	}
	if operation == OperationLowerIR && strings.TrimSpace(input.DSL) == "" && !hasJSON(input.IR) {
		return fmt.Errorf("lower-ir requires DSL or IR input")
	}
	if operation == OperationGenerate && !hasJSON(input.IR) {
		return fmt.Errorf("generate requires IR input")
	}
	return nil
}
func hasJSON(value json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(value))
	return trimmed != "" && trimmed != "null"
}
func validStatus(status Status) bool {
	return status == StatusPass || status == StatusFail || status == StatusDeferred || status == StatusNotRun
}
func knownOperation(operation Operation) bool {
	switch operation {
	case OperationParseAST, OperationLowerIR, OperationGenerate, OperationLiftBX,
		OperationResolveLSP, OperationCacheKey, OperationEmitEvidence, OperationCompareEvidence:
		return true
	default:
		return false
	}
}
