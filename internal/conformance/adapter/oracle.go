package adapter

import "fmt"

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
	Matched     bool
	ExitCode    ExitCode
	Detail      string
	FailureCode string
}

// Evaluate applies expected-negative and deferred semantics after validation.
func Evaluate(request Request, response Response) Evaluation {
	if err := request.Validate(); err != nil {
		return Evaluation{ExitCode: ExitRunnerError, Detail: err.Error()}
	}
	if err := response.validateIdentity(request); err != nil {
		return Evaluation{ExitCode: ExitCanonicalError, Detail: err.Error()}
	}
	if _, err := response.Normalized(); err != nil {
		return Evaluation{ExitCode: ExitCanonicalError, Detail: err.Error()}
	}
	if response.Status == StatusDeferred || response.Status == StatusNotRun {
		if request.Expected.Status == response.Status {
			return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "expected unavailable result"}
		}
		return Evaluation{ExitCode: ExitDeferred, Detail: fmt.Sprintf("observed %s", response.Status)}
	}
	if response.Status != request.Expected.Status {
		return Evaluation{ExitCode: ExitMismatch, Detail: fmt.Sprintf("expected %s, observed %s", request.Expected.Status, response.Status)}
	}
	if request.Expected.Status == StatusPass {
		return evaluatePass(response)
	}
	return evaluateFailure(request.Expected, response)
}

func (r Response) validateIdentity(request Request) error {
	if r.Schema != ProtocolSchema || r.Fixture != request.Fixture || r.Operation != request.Operation {
		return fmt.Errorf("response identity does not match request")
	}
	return nil
}

func evaluatePass(response Response) Evaluation {
	if response.Failure != nil {
		return Evaluation{ExitCode: ExitMismatch, Detail: "pass response contains failure details"}
	}
	if !response.PromotionEligible {
		return Evaluation{ExitCode: ExitMismatch, Detail: "pass response is not promotion eligible"}
	}
	return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "pass accepted"}
}

func evaluateFailure(expected Expectation, response Response) Evaluation {
	if response.Failure == nil {
		return Evaluation{ExitCode: ExitMismatch, Detail: "negative result omitted failure details"}
	}
	if !response.Failure.NoWrite || !response.Measurements.NoWrite {
		return Evaluation{ExitCode: ExitMismatch, Detail: "negative result did not prove no-write"}
	}
	if expected.FailureCode != "" && expected.FailureCode != response.Failure.Code {
		return Evaluation{
			ExitCode:    ExitMismatch,
			Detail:      fmt.Sprintf("expected failure %s, observed %s", expected.FailureCode, response.Failure.Code),
			FailureCode: response.Failure.Code,
		}
	}
	return Evaluation{Matched: true, ExitCode: ExitOK, Detail: "expected negative accepted", FailureCode: response.Failure.Code}
}
