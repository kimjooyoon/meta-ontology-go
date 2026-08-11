package performance

import "fmt"

// Host identifies the implementation that produced evidence.
type Host string

const (
	GoHosted   Host = "go-hosted"
	GoooHosted Host = "gooo-hosted"

	OperationsPerIterationUnit  = "operations/iteration"
	AllocationsPerIterationUnit = "allocations/iteration"
)

var standardStageNames = [...]string{"parser", "semantic", "query", "generator", "cache"}

// Status identifies the verification state of an evidence record.
type Status string

const (
	StatusVerified    Status = "verified"
	StatusFailed      Status = "failed"
	StatusPlanned     Status = "planned"
	StatusUnavailable Status = "unavailable"
)

// Contract fixes the identity and units that every host must report.
type Contract struct {
	ID             string
	Stage          string
	OperationUnit  string
	AllocationUnit string
}

// StandardContracts returns contracts for the five compiler stages in canonical order.
func StandardContracts() []Contract {
	contracts := make([]Contract, len(standardStageNames))
	for i, stage := range standardStageNames {
		contracts[i] = Contract{
			ID:             "performance://stage/" + stage,
			Stage:          stage,
			OperationUnit:  OperationsPerIterationUnit,
			AllocationUnit: AllocationsPerIterationUnit,
		}
	}
	return contracts
}

// Validate checks the stable identity and comparable metric units.
func (c Contract) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("performance contract ID is empty")
	}
	if c.Stage == "" {
		return fmt.Errorf("performance contract %q has no stage", c.ID)
	}
	if c.OperationUnit == "" || c.AllocationUnit == "" {
		return fmt.Errorf("performance contract %q has incomplete metric units", c.ID)
	}
	return nil
}

func (h Host) valid() bool {
	return h == GoHosted || h == GoooHosted
}

func (s Status) valid() bool {
	switch s {
	case StatusVerified, StatusFailed, StatusPlanned, StatusUnavailable:
		return true
	default:
		return false
	}
}
