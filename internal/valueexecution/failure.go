package valueexecution

import (
	"errors"
	"fmt"
)

const (
	ReasonSourceReadFailed          = "VALUE_SOURCE_READ_FAILED"
	ReasonSourceParseFailed         = "VALUE_SOURCE_PARSE_FAILED"
	ReasonSemanticBindingFailed     = "VALUE_SEMANTIC_BINDING_FAILED"
	ReasonPlanRequired              = "VALUE_PLAN_REQUIRED"
	ReasonActivityNotFound          = "VALUE_ACTIVITY_NOT_FOUND"
	ReasonProgramMissing            = "VALUE_PROGRAM_MISSING"
	ReasonProgramUnknown            = "VALUE_PROGRAM_UNKNOWN"
	ReasonOperandInvalid            = "VALUE_OPERAND_INVALID"
	ReasonSignatureArityUnsupported = "VALUE_SIGNATURE_ARITY_UNSUPPORTED"
	ReasonSignatureTypeMismatch     = "VALUE_SIGNATURE_TYPE_MISMATCH"
	ReasonOperationSpecInvalid      = "VALUE_OPERATION_SPEC_INVALID"
	ReasonOperationIRInvalid        = "VALUE_OPERATION_IR_INVALID"
	ReasonInputArityMismatch        = "VALUE_INPUT_ARITY_MISMATCH"
	ReasonIntegerOverflow           = "VALUE_INTEGER_OVERFLOW"
	ReasonIndicatorUnsatisfied      = "VALUE_INDICATOR_UNSATISFIED"
	ReasonProgramAuthorityInvalid   = "VALUE_PROGRAM_AUTHORITY_INVALID"
	ReasonProgramAuthorityMismatch  = "VALUE_PROGRAM_AUTHORITY_MISMATCH"
	ReasonResultHandleInvalid       = "VALUE_RESULT_HANDLE_INVALID"
	ReasonResultProducerMismatch    = "VALUE_RESULT_PRODUCER_MISMATCH"
	ReasonPlanInvalid               = "VALUE_PLAN_INVALID"
	ReasonExternalInputMissing      = "VALUE_EXTERNAL_INPUT_MISSING"
	ReasonExternalInputUnexpected   = "VALUE_EXTERNAL_INPUT_UNEXPECTED"
	ReasonBindingResultInvalid      = "VALUE_BINDING_RESULT_INVALID"
	ReasonPlanExecutionFailed       = "VALUE_PLAN_EXECUTION_FAILED"
)

type Failure struct {
	Code   string
	Stage  string
	Step   string
	Detail string
}

func (failure Failure) Error() string {
	if failure.Detail == "" {
		return failure.Code
	}
	return fmt.Sprintf("%s: %s", failure.Code, failure.Detail)
}

func failAt(code, stage, step, detail string) error {
	return Failure{Code: code, Stage: stage, Step: step, Detail: detail}
}

func FailureOf(err error) (Failure, bool) {
	if err == nil {
		return Failure{}, false
	}
	return errors.AsType[Failure](err)
}

func Reason(err error) string {
	if err == nil {
		return ""
	}
	if failure, ok := FailureOf(err); ok {
		return failure.Code
	}
	return ReasonSemanticBindingFailed
}
