package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Requirement names one finite, ordered path. PathID is the stable identity
// used in the result sets; names and aliases are not accepted as substitutes.
type Requirement struct {
	PathID        semantic.ID
	RecordIDs     []semantic.ID
	ExpectedKinds []semantic.InferenceKind
	StartID       semantic.ID
	EndID         semantic.ID
}

// Status is the complete decision vocabulary for path-closure evaluation.
type Status string

const (
	PASS        Status = "PASS"
	FAIL_CLOSED Status = "FAIL_CLOSED"
	UNKNOWN     Status = "UNKNOWN"

	StatusPass       = PASS
	StatusFailClosed = FAIL_CLOSED
	StatusUnknown    = UNKNOWN
)

// Stable result codes. The code describes the first applicable result class
// under the fixed precedence used by Evaluate; it carries no proof claim.
const (
	CodePass            = "PATH_CLOSURE_V1_PASS"
	CodeDuplicate       = "PATH_CLOSURE_V1_DUPLICATE"
	CodeMalformed       = "PATH_CLOSURE_V1_MALFORMED"
	CodeMissingRecord   = "PATH_CLOSURE_V1_MISSING_RECORD"
	CodeMissingEvidence = "PATH_CLOSURE_V1_MISSING_EVIDENCE"
	CodeMissingSnapshot = "PATH_CLOSURE_V1_MISSING_SNAPSHOT"
	CodeZeroDenominator = "PATH_CLOSURE_V1_ZERO_DENOMINATOR"
	CodeInvalidSemantic = "PATH_CLOSURE_V1_INVALID_SEMANTIC_PATH"
)

// Result is a deterministic coverage result over the named requirements.
// Every ID slice is sorted and contains each ID at most once.
type Result struct {
	Required    []semantic.ID
	Complete    []semantic.ID
	Missing     []semantic.ID
	Malformed   []semantic.ID
	Duplicate   []semantic.ID
	Numerator   int
	Denominator int
	Status      Status
	Code        string
}

// Evaluation is an alternate descriptive name for Result.
type Evaluation = Result
type requirementState struct {
	raw        Requirement
	normalized Requirement
	duplicate  bool
	malformed  bool
}
type issueClass uint8

const (
	issueMalformed issueClass = iota + 1
	issueDuplicate
	issueMissingEvidence
	issueMissingSnapshot
)
