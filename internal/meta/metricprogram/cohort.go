package metricprogram

// CohortTerminalSpec declares an expected subject independently of observed bindings.
type CohortTerminalSpec struct {
	Subject  string
	Carrier  string
	Family   string
	Trilemma string
}

// CanonicalCohortTerminals returns the terminals declared by the executable Gooo source.
func CanonicalCohortTerminals() []CohortTerminalSpec {
	return sourceCohortTerminals()
}
