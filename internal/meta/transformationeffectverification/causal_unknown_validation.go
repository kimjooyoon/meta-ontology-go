package transformationeffectverification

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func verificationFailureIndex(failures []generation.ObservationFailure) (map[string]generation.ObservationFailure, error) {
	index := make(map[string]generation.ObservationFailure, len(failures))
	for _, failure := range failures {
		if failure.ActionIndicatorID == "" || failure.Decision != string(generation.ReceiptDecisionRefuted) ||
			failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			failure.NextOperation == "" || failure.BlockedBy == nil {
			return nil, fmt.Errorf("failure context is incomplete")
		}
		if _, exists := index[failure.ActionIndicatorID]; exists {
			return nil, fmt.Errorf("failure context is duplicated")
		}
		index[failure.ActionIndicatorID] = failure
	}
	return index, nil
}

func validateVerificationCausalUnknown(unknown generation.ReceiptUnknown) error {
	if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" || unknown.Operation == "" ||
		unknown.Activity == "" || unknown.Output == "" || unknown.Executor == "" || unknown.Evaluator == "" ||
		unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.NextOperation == "" ||
		unknown.BlockedBy == nil {
		return fmt.Errorf("unknown context is incomplete")
	}
	return nil
}

func verificationCausalUnknownRecord(unknown generation.ReceiptUnknown) transformationeffect.CausalUnknownRecord {
	return transformationeffect.CausalUnknownRecord{
		ActionIndicatorID: unknown.ActionIndicatorID, RequiredIndicatorID: unknown.RequiredIndicatorID,
		Stage: unknown.Stage, Step: unknown.Step, Reason: string(unknown.Reason),
		UnknownClass: unknown.UnknownClass, NextOperation: unknown.NextOperation,
		BlockedBy: append([]string{}, unknown.BlockedBy...)}
}

func causalUnknownKey(record transformationeffect.CausalUnknownRecord) string {
	return strings.Join([]string{record.ActionIndicatorID, record.RequiredIndicatorID,
		record.Stage, record.Step, record.Reason, record.UnknownClass, record.NextOperation,
		strings.Join(record.BlockedBy, "\x00")}, "\x00")
}
