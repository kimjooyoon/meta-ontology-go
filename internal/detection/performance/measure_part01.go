package performance

import (
	"fmt"
)

// Config controls deterministic observation counts.
type Config struct {
	Iterations     int
	AllocationRuns int
}

// DefaultConfig provides a small repeatable sample suitable for unit tests.
func DefaultConfig() Config {
	return Config{Iterations: 16, AllocationRuns: 8}
}
func (c Config) normalized() (Config, error) {
	defaults := DefaultConfig()
	if c.Iterations == 0 {
		c.Iterations = defaults.Iterations
	}
	if c.AllocationRuns == 0 {
		c.AllocationRuns = defaults.AllocationRuns
	}
	if c.Iterations < 1 {
		return Config{}, fmt.Errorf("performance iterations must be positive")
	}
	if c.AllocationRuns < 1 {
		return Config{}, fmt.Errorf("performance allocation runs must be positive")
	}
	return c, nil
}

// Observation contains deterministic counts for one stage.
type Observation struct {
	Stage                   Stage
	Iterations              int
	Operations              uint64
	AllocationsPerIteration float64
	Budget                  Budget
}

func (o Observation) operationsPerIteration() float64 {
	if o.Iterations == 0 {
		return 0
	}
	return float64(o.Operations) / float64(o.Iterations)
}

// OperationsPerIteration returns the measured operation count per invocation.
func (o Observation) OperationsPerIteration() float64 {
	return o.operationsPerIteration()
}
