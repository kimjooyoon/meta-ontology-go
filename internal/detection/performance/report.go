package performance

import (
	"fmt"
	"strings"
)

// Report contains observations and all detected budget overruns.
type Report struct {
	Observations []Observation
	Violations   []Violation
}

// Passed reports whether every observed stage stayed within its configured
// deterministic operation and allocation budgets.
func (r Report) Passed() bool {
	return len(r.Violations) == 0
}

// String renders a stable, wall-clock-free report for CI or local diagnostics.
func (r Report) String() string {
	var output strings.Builder
	output.WriteString("performance report\n")
	for _, observation := range r.Observations {
		fmt.Fprintf(&output,
			"stage=%s iterations=%d operations=%d operations/iteration=%.2f allocations/iteration=%.2f\n",
			observation.Stage,
			observation.Iterations,
			observation.Operations,
			observation.OperationsPerIteration(),
			observation.AllocationsPerIteration,
		)
	}
	if r.Passed() {
		output.WriteString("status=pass\n")
		return output.String()
	}
	output.WriteString("status=budget-overrun\n")
	for _, violation := range r.Violations {
		fmt.Fprintf(&output, "violation stage=%s metric=%s actual=%.2f limit=%.2f\n",
			violation.Stage, violation.Metric, violation.Actual, violation.Limit)
	}
	return output.String()
}
