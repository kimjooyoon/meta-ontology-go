package pressureshadow

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionFailClosed Decision = "FAIL_CLOSED"
)

type Reason string

const (
	ReasonNone                 Reason = "NONE"
	ReasonInvalidInput         Reason = "INVALID_INPUT"
	ReasonRequiredInputMissing Reason = "REQUIRED_INPUT_MISSING"
	ReasonMissingPathCoverage  Reason = "MISSING_PATH_COVERAGE"
	ReasonOrphanPathCoverage   Reason = "ORPHAN_PATH_COVERAGE"
	ReasonBindingMismatch      Reason = "BINDING_MISMATCH"
)

type EnforcementEffect string

const EnforcementNoEffect EnforcementEffect = "NO_EFFECT"

type Result struct {
	Schema                 string            `json:"schema"`
	InputDigest            string            `json:"input_digest"`
	Decision               Decision          `json:"decision"`
	Reason                 Reason            `json:"reason"`
	MissingPathIDs         []string          `json:"missing_path_ids"`
	OrphanPathIDs          []string          `json:"orphan_path_ids"`
	MissingBindingPathIDs  []string          `json:"missing_binding_path_ids"`
	BindingMismatchPathIDs []string          `json:"binding_mismatch_path_ids"`
	EnforcementEffect      EnforcementEffect `json:"enforcement_effect"`
	ResultDigest           string            `json:"result_digest"`
	ReplayDigest           string            `json:"replay_digest"`
}

// Validate checks only the S1a2a path-row and outer tuple contract.
func Validate(input Input) Result {
	canonical, err := CanonicalInputBytes(input)
	if err != nil {
		return makeResult(CanonicalInputDigest(input), DecisionFailClosed, ReasonInvalidInput,
			nil, nil, nil, nil)
	}
	inputDigest := digestBytes(canonical)
	paths := selectorPathIDs(input)
	rows := coverageRows(input)
	missing := missingPathIDs(paths, rows)
	orphan := orphanPathIDs(paths, rows)
	missingBinding, mismatch := bindingIssues(input, paths, rows)
	decision, reason := pathDecision(len(paths), missing, orphan, missingBinding, mismatch)
	return makeResult(inputDigest, decision, reason, missing, orphan, missingBinding, mismatch)
}

// ValidateBytes is the strict raw-wire boundary for S1a2a validation.
func ValidateBytes(data []byte) Result {
	input, err := DecodeInput(data)
	if err != nil {
		return makeResult(invalidInputDigest(data), DecisionFailClosed, ReasonInvalidInput,
			nil, nil, nil, nil)
	}
	return Validate(input)
}
func invalidInputDigest(data []byte) string {
	return digestBytes(append([]byte("invalid-input\x00"), data...))
}
