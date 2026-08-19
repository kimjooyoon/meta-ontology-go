package performance

import (
	"fmt"
	"testing"
)

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
