package sourceexecution

import (
	"fmt"
	"strings"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("SOURCE_EXECUTION_SCHEMA_UNKNOWN")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("SOURCE_EXECUTION_DECISION_UNKNOWN")
	}
	if receipt.Resolution != "EXACT" {
		return fmt.Errorf("SOURCE_EXECUTION_RESOLUTION_UNKNOWN")
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return fmt.Errorf("SOURCE_EXECUTION_EFFECT_BOUNDARY_VIOLATED")
	}
	if receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("SOURCE_EXECUTION_DIGEST_MISMATCH")
	}
	if !strings.HasPrefix(receipt.SourceDigest, "sha256:") {
		return fmt.Errorf("SOURCE_EXECUTION_SOURCE_UNBOUND")
	}
	if receipt.Decision == "PASS" {
		return validateSuccess(receipt)
	}
	if len(receipt.Diagnostics) != 1 || receipt.Diagnostics[0].Code != receipt.Reason || len(receipt.Events) != 0 {
		return fmt.Errorf("SOURCE_EXECUTION_FAILURE_DIAGNOSTIC_INVALID")
	}
	return nil
}

func validateSuccess(receipt Receipt) error {
	want := []string{"SOURCE_PARSED", "SEMANTIC_LOWERED", "ACTIVITY_INVOKED", "ENTITY_PRODUCED"}
	wantSubjects := []string{receipt.SourceDigest, receipt.SemanticDigest, receipt.Entry.Activity, receipt.Entry.Output.ID}
	if receipt.Reason != "SOURCE_ACTIVITY_EXECUTED" || len(receipt.Diagnostics) != 0 ||
		receipt.Entry.Activity == "" || receipt.Entry.Output.ID == "" || len(receipt.Events) != len(want) {
		return fmt.Errorf("SOURCE_EXECUTION_SUCCESS_INVALID")
	}
	for index, event := range receipt.Events {
		if event.Sequence != index+1 || event.Kind != want[index] || event.Subject == "" {
			return fmt.Errorf("SOURCE_EXECUTION_EVENT_INVALID")
		}
		if event.Subject != wantSubjects[index] {
			return fmt.Errorf("SOURCE_EXECUTION_EVENT_SUBJECT_BINDING_INVALID")
		}
	}
	return nil
}
