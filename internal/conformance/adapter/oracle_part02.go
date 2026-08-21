package adapter

import (
	"fmt"
)

func evaluatePass(response Response, request Request, observation *NoWriteObservation) Evaluation {
	if response.Failure != nil {
		return Evaluation{ExitCode: ExitMismatch, Detail: "pass response contains failure details"}
	}
	if response.ProducerClaims.NoWrite != nil || observation != nil {
		if observation == nil {
			return oracleFailure(OraclePASS001, "producer-only no-write claim has no observer proof", "")
		}
		if err := observation.VerifyNoWrite(request); err != nil {
			return oracleFailure(OraclePASS001, err.Error(), "")
		}
	}
	if !response.PromotionEligible {
		return Evaluation{ExitCode: ExitMismatch, Detail: "pass response is not promotion eligible"}
	}
	return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "pass accepted", PromotionEligible: true}
}
func evaluateFailure(request Request, response Response, observation *NoWriteObservation) Evaluation {
	if response.Failure == nil {
		return oracleFailure(OracleFAIL001, "negative result omitted failure details", "")
	}
	if request.Expected.FailureCode != "" && request.Expected.FailureCode != response.Failure.Code {
		return oracleFailure(OracleFAIL002, fmt.Sprintf("expected failure %s, observed %s", request.Expected.FailureCode, response.Failure.Code), response.Failure.Code)
	}
	if observation == nil {
		return oracleFailure(OracleNW001, "observer evidence is required for negative results", response.Failure.Code)
	}
	if err := observation.VerifyNoWrite(request); err != nil {
		return oracleFailureFromError(err, response.Failure.Code)
	}
	return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "expected negative accepted", FailureCode: response.Failure.Code}
}
func oracleFailure(code, detail, failureCode string) Evaluation {
	return Evaluation{ExitCode: ExitMismatch, Detail: code + ": " + detail, FailureCode: failureCode, OracleCode: code}
}
func canonicalOracleFailure(code, detail string) Evaluation {
	return Evaluation{ExitCode: ExitCanonicalError, Detail: code + ": " + detail, OracleCode: code}
}
func oracleFailureFromError(err error, failureCode string) Evaluation {
	if evidenceErr, ok := err.(NoWriteEvidenceError); ok {
		return oracleFailure(evidenceErr.Code, evidenceErr.Detail, failureCode)
	}
	return oracleFailure(OracleNW003, err.Error(), failureCode)
}
