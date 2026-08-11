package performance

import "fmt"

// Comparison records whether both hosts produced evidence that can be compared.
// A non-comparable result is an explicit pending state, not a successful check.
type Comparison struct {
	Contract        Contract
	GoEvidence      Evidence
	GoooEvidence    Evidence
	Comparable      bool
	Reason          string
	OperationDelta  float64
	AllocationDelta float64
}

// Compare validates a shared contract and returns a comparison even when one
// host is planned, unavailable, or failed. Such states never become comparable.
func Compare(contract Contract, goEvidence, goooEvidence Evidence) (Comparison, error) {
	if err := contract.Validate(); err != nil {
		return Comparison{}, err
	}
	if err := goEvidence.Validate(contract); err != nil {
		return Comparison{}, fmt.Errorf("go-hosted evidence: %w", err)
	}
	if err := goooEvidence.Validate(contract); err != nil {
		return Comparison{}, fmt.Errorf("gooo-hosted evidence: %w", err)
	}
	if goEvidence.Host != GoHosted || goooEvidence.Host != GoooHosted {
		return Comparison{}, fmt.Errorf("evidence hosts must be %q and %q", GoHosted, GoooHosted)
	}
	comparison := Comparison{
		Contract:     contract,
		GoEvidence:   goEvidence,
		GoooEvidence: goooEvidence,
	}
	if goEvidence.Status != StatusVerified {
		comparison.Reason = fmt.Sprintf("go-hosted evidence is %s", goEvidence.Status)
		return comparison, nil
	}
	if goooEvidence.Status != StatusVerified {
		comparison.Reason = fmt.Sprintf("gooo-hosted evidence is %s", goooEvidence.Status)
		return comparison, nil
	}
	comparison.Comparable = true
	comparison.Reason = "both hosts have verified evidence"
	comparison.OperationDelta = float64(goooEvidence.OperationsPerIteration) -
		float64(goEvidence.OperationsPerIteration)
	comparison.AllocationDelta = float64(goooEvidence.AllocationsPerIteration) -
		float64(goEvidence.AllocationsPerIteration)
	return comparison, nil
}
