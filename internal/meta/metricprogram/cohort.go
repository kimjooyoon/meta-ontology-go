package metricprogram

// CohortTerminalSpec declares an expected subject independently of observed bindings.
type CohortTerminalSpec struct {
	Subject  string `gooo:"gooo://metric-registry/CohortTerminalDeclaration/Subject"`
	Carrier  string `gooo:"gooo://metric-registry/CohortTerminalDeclaration/Carrier"`
	Family   string `gooo:"gooo://metric-registry/CohortTerminalDeclaration/Family"`
	Trilemma string `gooo:"gooo://metric-registry/CohortTerminalDeclaration/Trilemma"`
}

// CanonicalCohortTerminals returns the terminals declared by the executable Gooo source.
func CanonicalCohortTerminals() []CohortTerminalSpec {
	return sourceCohortTerminals()
}
