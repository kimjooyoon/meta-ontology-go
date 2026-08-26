package selectiveci

import (
	"fmt"
)

type validationError struct {
	reason string
	err    error
}

func (e *validationError) Error() string { return e.reason + ": " + e.err.Error() }
func failure(reason, message string) error {
	return &validationError{reason: reason, err: fmt.Errorf("%s", message)}
}
func reasonFor(err error) string {
	if typed, ok := err.(*validationError); ok {
		return typed.reason
	}
	return ReasonInvalidInput
}
func (input Input) Validate() error {
	if input.SchemaVersion != SchemaVersion {
		return failure(ReasonUnsupportedSchema, "unsupported schema_version")
	}
	if err := input.Base.Validate(); err != nil {
		return err
	}
	if err := input.Head.Validate(); err != nil {
		return err
	}
	if input.CPUCapacity == 0 {
		return failure(ReasonInvalidInput, "cpu_capacity must be positive")
	}
	if input.Receipts == nil || input.ProvenancePaths == nil {
		return failure(ReasonInvalidInput, "resource_receipts and provenance_paths are required")
	}
	if err := input.Registry.Validate(); err != nil {
		return err
	}
	if err := validateReceipts(input.Receipts); err != nil {
		return err
	}
	return validateProvenancePaths(input.ProvenancePaths)
}
