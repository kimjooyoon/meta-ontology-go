package transformationeffect

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func validateCausalUnknown(unknown generation.ReceiptUnknown) error {
	if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" || unknown.Operation == "" ||
		unknown.Activity == "" || unknown.Output == "" || unknown.Executor == "" || unknown.Evaluator == "" ||
		unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.NextOperation == "" ||
		unknown.BlockedBy == nil {
		return fmt.Errorf("causal unknown is missing required fields")
	}
	return nil
}

func causalUnknownKey(record CausalUnknownRecord) string {
	return strings.Join([]string{record.ActionIndicatorID, record.RequiredIndicatorID,
		record.Stage, record.Step, record.Reason, record.UnknownClass, record.NextOperation,
		strings.Join(record.BlockedBy, "\x00")}, "\x00")
}
