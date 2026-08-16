package performance

import (
	"fmt"
	"testing"
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

// Measure executes a stage for a fixed number of iterations and samples its
// allocation count separately. The runner should therefore be repeatable.
func Measure(spec StageSpec, config Config) (Observation, error) {
	if err := spec.validate(); err != nil {
		return Observation{}, err
	}
	config, err := config.normalized()
	if err != nil {
		return Observation{}, err
	}
	operations, err := measureOperations(spec.Run, config.Iterations)
	if err != nil {
		return Observation{}, fmt.Errorf("measure %q: %w", spec.Stage, err)
	}
	allocations, err := measureAllocations(spec.Run, config.AllocationRuns)
	if err != nil {
		return Observation{}, fmt.Errorf("measure allocations for %q: %w", spec.Stage, err)
	}
	return Observation{
		Stage:                   spec.Stage,
		Iterations:              config.Iterations,
		Operations:              operations,
		AllocationsPerIteration: allocations,
		Budget:                  spec.Budget,
	}, nil
}

func measureOperations(run Runner, iterations int) (uint64, error) {
	var total uint64
	var counter Counter
	for i := 0; i < iterations; i++ {
		counter.reset()
		if err := run(&counter); err != nil {
			return 0, err
		}
		total += counter.Operations()
	}
	return total, nil
}

func measureAllocations(run Runner, runs int) (float64, error) {
	var runErr error
	var counter Counter
	allocations := testing.AllocsPerRun(runs, func() {
		counter.reset()
		if err := run(&counter); err != nil && runErr == nil {
			runErr = err
		}
	})
	return allocations, runErr
}

// MeasureAll observes all supplied stages in canonical pipeline order and
// evaluates each stage against its own budget. The input slice is not changed.
func MeasureAll(specs []StageSpec, config Config) (Report, error) {
	if err := validateSpecs(specs); err != nil {
		return Report{}, err
	}
	ordered := orderedSpecs(specs)
	report := Report{Observations: make([]Observation, 0, len(ordered))}
	for _, spec := range ordered {
		observation, err := Measure(spec, config)
		if err != nil {
			return report, err
		}
		report.Observations = append(report.Observations, observation)
		report.Violations = append(report.Violations, DetectBudget(observation)...)
	}
	return report, nil
}
