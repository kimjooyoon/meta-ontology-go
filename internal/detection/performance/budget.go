package performance

// Budget contains per-iteration limits. A zero limit disables that metric so a
// caller can constrain only the dimensions it has calibrated.
type Budget struct {
	MaxOperationsPerIteration uint64
	MaxAllocsPerIteration     uint64
}

// Metric identifies a budget dimension.
type Metric string

const (
	OperationsMetric Metric = "operations/iteration"
	AllocsMetric     Metric = "allocations/iteration"
)

// Violation describes one observed metric above its configured limit.
type Violation struct {
	Stage  Stage
	Metric Metric
	Actual float64
	Limit  float64
}

// DetectBudget returns all configured budget overruns for one observation.
func DetectBudget(observation Observation) []Violation {
	violations := make([]Violation, 0, 2)
	operations := observation.operationsPerIteration()
	if limit := observation.Budget.MaxOperationsPerIteration; limit > 0 && operations > float64(limit) {
		violations = append(violations, Violation{
			Stage:  observation.Stage,
			Metric: OperationsMetric,
			Actual: operations,
			Limit:  float64(limit),
		})
	}
	if limit := observation.Budget.MaxAllocsPerIteration; limit > 0 &&
		observation.AllocationsPerIteration > float64(limit) {
		violations = append(violations, Violation{
			Stage:  observation.Stage,
			Metric: AllocsMetric,
			Actual: observation.AllocationsPerIteration,
			Limit:  float64(limit),
		})
	}
	return violations
}
