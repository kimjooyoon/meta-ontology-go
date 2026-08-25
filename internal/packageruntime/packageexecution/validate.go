package packageexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("packageexecution: schema mismatch")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("packageexecution: unknown decision %q", receipt.Decision)
	}
	if receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" {
		return fmt.Errorf("packageexecution: unknown resolution %q", receipt.Resolution)
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return fmt.Errorf("packageexecution: execution acquired mutation effects")
	}
	if err := validateSources(receipt.Sources); err != nil {
		return err
	}
	if err := validateEvents(receipt.Events); err != nil {
		return err
	}
	if err := validateExecution(receipt); err != nil {
		return err
	}
	want := receipt.Digest
	seal(&receipt)
	if want == "" || receipt.Digest != want {
		return fmt.Errorf("packageexecution: receipt digest mismatch")
	}
	return nil
}

func validateEvents(values []Event) error {
	for index, event := range values {
		if event.Sequence != index+1 || event.Kind == "" || event.Subject == "" {
			return fmt.Errorf("packageexecution: invalid event sequence")
		}
	}
	return nil
}

func validateExecution(receipt Receipt) error {
	if receipt.Execution != nil {
		if err := sourceexecution.Validate(*receipt.Execution); err != nil {
			return fmt.Errorf("packageexecution: nested receipt: %w", err)
		}
	}
	if receipt.Decision != "PASS" {
		return nil
	}
	if receipt.Resolution != "EXACT" || len(receipt.Sources) < 2 || receipt.Execution == nil {
		return fmt.Errorf("packageexecution: passing receipt is incomplete")
	}
	if receipt.Execution.Decision != "PASS" || receipt.Reason != "PACKAGE_EXECUTED" || len(receipt.Diagnostics) != 0 {
		return fmt.Errorf("packageexecution: passing receipt contradicts nested execution")
	}
	return nil
}
