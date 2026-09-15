package metainvocation

type PlannedCheck struct {
	ID      string         `json:"id"`
	Command string         `json:"command"`
	Files   []string       `json:"files"`
	Reasons []RuleEvidence `json:"reasons"`
}

type CheckPlan struct {
	Schema      string         `json:"schema"`
	CaseID      string         `json:"case_id"`
	InputDigest string         `json:"input_digest"`
	Checks      []PlannedCheck `json:"checks"`
	Digest      string         `json:"digest"`
}

type UnknownCause struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
	File   string `json:"file,omitempty"`
}
