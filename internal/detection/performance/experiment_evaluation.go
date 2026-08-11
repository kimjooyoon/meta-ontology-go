package performance

import "fmt"

// RunEvidence is the host output consumed by the experiment evaluator.
type RunEvidence struct {
	Host                    Host   `json:"host"`
	Status                  Status `json:"status"`
	FixtureID               string `json:"fixture_id"`
	InputDigest             string `json:"input_digest"`
	Repetitions             uint64 `json:"repetitions"`
	OperationsPerIteration  uint64 `json:"operations_per_iteration"`
	AllocationsPerIteration uint64 `json:"allocations_per_iteration"`
	Source                  string `json:"source"`
}

// Result is the explicit pass, fail, or deferred decision for one host.
type Result struct {
	HypothesisID string  `json:"hypothesis_id"`
	Host         Host    `json:"host"`
	Outcome      Outcome `json:"outcome"`
	Reason       string  `json:"reason"`
	Operations   uint64  `json:"operations"`
	Allocations  uint64  `json:"allocations"`
}

// ComparisonResult combines both host decisions and their deterministic delta.
type ComparisonResult struct {
	HypothesisID    string  `json:"hypothesis_id"`
	Go              Result  `json:"go"`
	Gooo            Result  `json:"gooo"`
	Outcome         Outcome `json:"outcome"`
	Reason          string  `json:"reason"`
	OperationDelta  float64 `json:"operation_delta"`
	AllocationDelta float64 `json:"allocation_delta"`
}

// Evaluate applies exact metrics and explicit status criteria to one host.
func Evaluate(hypothesis Hypothesis, evidence RunEvidence) (Result, error) {
	if err := hypothesis.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateRun(hypothesis, evidence); err != nil {
		return Result{}, err
	}
	result := Result{HypothesisID: hypothesis.ID, Host: evidence.Host,
		Operations: evidence.OperationsPerIteration, Allocations: evidence.AllocationsPerIteration}
	if evidence.Status == StatusPlanned || evidence.Status == StatusUnavailable {
		result.Outcome = OutcomeDeferred
		result.Reason = fmt.Sprintf("host status is %s", evidence.Status)
		return result, nil
	}
	if evidence.Status == StatusFailed {
		result.Outcome = OutcomeFail
		result.Reason = "host reported failed execution"
		return result, nil
	}
	if evidence.Repetitions != hypothesis.Repetitions {
		return failResult(result, "repetition count does not match hypothesis"), nil
	}
	if evidence.OperationsPerIteration != hypothesis.ExpectedOperations {
		return failResult(result, "operation count falsifies hypothesis"), nil
	}
	if evidence.AllocationsPerIteration != hypothesis.ExpectedAllocations {
		return failResult(result, "allocation count falsifies hypothesis"), nil
	}
	result.Outcome = OutcomePass
	result.Reason = "exact deterministic measurements match hypothesis"
	return result, nil
}

// CompareEvidence evaluates both hosts without treating deferred work as pass.
func CompareEvidence(hypothesis Hypothesis, goEvidence, goooEvidence RunEvidence) (ComparisonResult, error) {
	if goEvidence.Host != GoHosted || goooEvidence.Host != GoooHosted {
		return ComparisonResult{}, fmt.Errorf("evidence hosts must be %q and %q", GoHosted, GoooHosted)
	}
	goResult, err := Evaluate(hypothesis, goEvidence)
	if err != nil {
		return ComparisonResult{}, fmt.Errorf("go-hosted evaluation: %w", err)
	}
	goooResult, err := Evaluate(hypothesis, goooEvidence)
	if err != nil {
		return ComparisonResult{}, fmt.Errorf("gooo-hosted evaluation: %w", err)
	}
	comparison := ComparisonResult{HypothesisID: hypothesis.ID, Go: goResult, Gooo: goooResult,
		OperationDelta:  float64(goooResult.Operations) - float64(goResult.Operations),
		AllocationDelta: float64(goooResult.Allocations) - float64(goResult.Allocations)}
	if goResult.Outcome == OutcomeFail || goooResult.Outcome == OutcomeFail {
		comparison.Outcome = OutcomeFail
		comparison.Reason = "at least one host falsified the hypothesis"
		return comparison, nil
	}
	if goResult.Outcome == OutcomeDeferred || goooResult.Outcome == OutcomeDeferred {
		comparison.Outcome = OutcomeDeferred
		comparison.Reason = "a host implementation is not yet verifiable"
		return comparison, nil
	}
	comparison.Outcome = OutcomePass
	comparison.Reason = "both hosts matched the deterministic hypothesis"
	return comparison, nil
}

func validateRun(hypothesis Hypothesis, evidence RunEvidence) error {
	if !evidence.Host.valid() || !evidence.Status.valid() {
		return fmt.Errorf("run evidence has invalid host or status")
	}
	if evidence.FixtureID != hypothesis.Fixture.ID || evidence.InputDigest != hypothesis.Fixture.InputDigest {
		return fmt.Errorf("run evidence fixture identity does not match hypothesis")
	}
	if evidence.Source == "" {
		return fmt.Errorf("run evidence has no source")
	}
	if evidence.Status == StatusPlanned || evidence.Status == StatusUnavailable {
		if evidence.Repetitions != 0 || evidence.OperationsPerIteration != 0 || evidence.AllocationsPerIteration != 0 {
			return fmt.Errorf("deferred run evidence cannot report measurements")
		}
	}
	return nil
}

func failResult(result Result, reason string) Result {
	result.Outcome = OutcomeFail
	result.Reason = reason
	return result
}
