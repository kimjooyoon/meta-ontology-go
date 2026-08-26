package valueexecution

import (
	"errors"
	"fmt"
)

const (
	ReasonSourceReadFailed           = "VALUE_SOURCE_READ_FAILED"
	ReasonSourceParseFailed          = "VALUE_SOURCE_PARSE_FAILED"
	ReasonSemanticBindingFailed      = "VALUE_SEMANTIC_BINDING_FAILED"
	ReasonActivityNotFound           = "VALUE_ACTIVITY_NOT_FOUND"
	ReasonProgramMissing             = "VALUE_PROGRAM_MISSING"
	ReasonProgramUnknown             = "VALUE_PROGRAM_UNKNOWN"
	ReasonOperandInvalid             = "VALUE_OPERAND_INVALID"
	ReasonSignatureArityUnsupported  = "VALUE_SIGNATURE_ARITY_UNSUPPORTED"
	ReasonInputArityMismatch         = "VALUE_INPUT_ARITY_MISMATCH"
	ReasonIntegerOverflow            = "VALUE_INTEGER_OVERFLOW"
)

type Failure struct {
	Code   string
	Detail string
}

func (failure Failure) Error() string {
	if failure.Detail == "" {
		return failure.Code
	}
	return fmt.Sprintf("%s: %s", failure.Code, failure.Detail)
}

func fail(code, detail string) error {
	return Failure{Code: code, Detail: detail}
}

func Reason(err error) string {
	if err == nil {
		return ""
	}
	var failure Failure
	if errors.As(err, &failure) {
		return failure.Code
	}
	return ReasonSemanticBindingFailed
}
