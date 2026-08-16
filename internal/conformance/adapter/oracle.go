package adapter

import (
	"fmt"
	"strings"
)

// ExitCode is stable process semantics for fixture runners.
type ExitCode int

const (
	ExitOK             ExitCode = 0
	ExitMismatch       ExitCode = 10
	ExitDeferred       ExitCode = 20
	ExitRunnerError    ExitCode = 30
	ExitCanonicalError ExitCode = 40
)

// Evaluation is the oracle result; a nonzero code is never promotion.
type Evaluation struct {
	Matched           bool
	ExitCode          ExitCode
	Detail            string
	FailureCode       string
	OracleCode        string
	PromotionEligible bool
}

// Evaluate never accepts producer-carried no-write claims as proof.
func Evaluate(request Request, response Response) Evaluation {
	return EvaluateObserved(request, response, nil)
}

// EvaluateObserved evaluates a response with an independently captured trace.
func EvaluateObserved(request Request, response Response, observation *NoWriteObservation) Evaluation {
	if strings.TrimSpace(request.RunID) == "" || response.validateIdentity(request) != nil {
		return canonicalOracleFailure(OracleID001, "request, response, and run identity do not match")
	}
	if err := request.Validate(); err != nil {
		return Evaluation{ExitCode: ExitRunnerError, Detail: err.Error()}
	}
	if request.Expected.Status == StatusFail && response.Status == StatusFail && response.Failure == nil {
		return oracleFailure(OracleFAIL001, "negative response omitted failure details", "")
	}
	normalized, err := response.Normalized()
	if err != nil {
		return Evaluation{ExitCode: ExitCanonicalError, Detail: err.Error()}
	}
	response = normalized
	if response.Status == StatusDeferred || response.Status == StatusNotRun {
		return evaluateUnavailable(request.Expected.Status, response.Status)
	}
	if response.Status != request.Expected.Status {
		return Evaluation{ExitCode: ExitMismatch, Detail: fmt.Sprintf("expected %s, observed %s", request.Expected.Status, response.Status)}
	}
	if request.Expected.Status == StatusPass {
		return evaluatePass(response, request, observation)
	}
	return evaluateFailure(request, response, observation)
}

func (r Response) validateIdentity(request Request) error {
	if r.Schema != ProtocolSchema || r.Fixture != request.Fixture || r.Operation != request.Operation || r.RunID != request.RunID {
		return fmt.Errorf("response identity does not match request")
	}
	return nil
}

func evaluateUnavailable(expected, observed Status) Evaluation {
	if expected == observed {
		return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "expected unavailable result"}
	}
	return Evaluation{ExitCode: ExitDeferred, Detail: fmt.Sprintf("observed %s", observed)}
}

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
