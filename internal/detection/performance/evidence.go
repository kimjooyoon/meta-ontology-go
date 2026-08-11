package performance

import "fmt"

// Evidence is an immutable observation record for one host and contract.
// Metrics are meaningful only for verified or failed executions.
type Evidence struct {
	ContractID              string
	Host                    Host
	Status                  Status
	OperationsPerIteration  uint64
	AllocationsPerIteration uint64
	Source                  string
}

// Validate checks provenance and prevents planned work from carrying success
// shaped metrics.
func (e Evidence) Validate(contract Contract) error {
	if err := contract.Validate(); err != nil {
		return err
	}
	if e.ContractID != contract.ID {
		return fmt.Errorf("evidence contract %q does not match %q", e.ContractID, contract.ID)
	}
	if !e.Host.valid() {
		return fmt.Errorf("evidence has unknown host %q", e.Host)
	}
	if !e.Status.valid() {
		return fmt.Errorf("evidence has unknown status %q", e.Status)
	}
	if e.Source == "" {
		return fmt.Errorf("%s evidence has no source", e.Host)
	}
	if e.Status == StatusPlanned || e.Status == StatusUnavailable {
		if e.OperationsPerIteration != 0 || e.AllocationsPerIteration != 0 {
			return fmt.Errorf("%s evidence cannot report metrics with status %q", e.Host, e.Status)
		}
	}
	return nil
}
