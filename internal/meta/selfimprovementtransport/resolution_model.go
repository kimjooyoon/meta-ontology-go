package selfimprovementtransport

const (
	ResolutionPolicySchema = "gooo/self-improvement-transport-resolution-policy/v1"
	ResolutionClosed       = "CLOSED"
	ResolutionUnknown      = "UNKNOWN"
	ResolutionRefuted      = "REFUTED"
	ResolutionCaseCount    = 10
)

type ResolutionCase struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	Stage        string `json:"stage"`
	Step         string `json:"step"`
	UnknownClass string `json:"unknown_class,omitempty"`
	Reason       string `json:"reason"`
}

type ResolutionPolicy struct {
	Schema                  string           `json:"schema"`
	States                  []string         `json:"states"`
	CausalFields            []string         `json:"causal_fields"`
	Transitions             []string         `json:"transitions"`
	ArtifactIdentity        []string         `json:"artifact_identity"`
	CaseDenominator         int              `json:"case_denominator"`
	ClosedCases             int              `json:"closed_cases"`
	UnknownCases            int              `json:"unknown_cases"`
	RefutedCases            int              `json:"refuted_cases"`
	Metrics                 map[string]int   `json:"metrics"`
	Cases                   []ResolutionCase `json:"cases"`
	RefutedDominatesUnknown bool             `json:"refuted_dominates_unknown"`
}
