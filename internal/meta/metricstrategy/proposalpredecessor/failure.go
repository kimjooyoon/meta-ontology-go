package proposalpredecessor

import (
	"errors"
	"fmt"
)

type Failure struct {
	Reason string
	Err    error
}

func (failure *Failure) Error() string {
	if failure.Err == nil {
		return failure.Reason
	}
	return fmt.Sprintf("%s: %v", failure.Reason, failure.Err)
}

func (failure *Failure) Unwrap() error { return failure.Err }

func FailureReason(err error) string {
	var failure *Failure
	if errors.As(err, &failure) {
		return failure.Reason
	}
	return ""
}

func KnownFailureReason(reason string) bool {
	switch reason {
	case ReasonNotFound, ReasonAmbiguous, ReasonEvidenceUnknown,
		ReasonJobCardinality, ReasonRunPaginationIncomplete,
		ReasonJobPaginationIncomplete, ReasonArtifactPaginationIncomplete,
		ReasonAPIUnavailable, ReasonAPIPermissionDenied, ReasonResponseMalformed,
		ReasonArtifactPayloadUnavailable, ReasonRedirectOriginMismatch:
		return true
	default:
		return false
	}
}
