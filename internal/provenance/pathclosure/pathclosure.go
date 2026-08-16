// Package pathclosure evaluates explicitly named semantic inference paths.
//
// It deliberately does not search the path graph. A requirement is complete
// only when the exact record and kind sequence named by the caller is valid.
package pathclosure

import (
	"errors"
	"fmt"

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

// Evaluate validates path semantics before evaluating any requirement. It
// indexes each normalized edge once and checks only the exact named sequence.
func Evaluate(path semantic.InferencePathV1, requirements []Requirement) Result {
	normalized, pathErr := path.Normalized()
	result := Result{Denominator: len(requirements)}
	states, required := prepareRequirements(requirements)
	result.Required = required
	if len(requirements) == 0 {
		result.Status, result.Code = UNKNOWN, CodeZeroDenominator
		return result
	}
	if pathErr != nil {
		return resultForSemanticError(result, states, pathErr)
	}

	edges, duplicateEdges := indexEdges(normalized.Edges)
	for _, state := range states {
		if state.duplicate {
			result.Duplicate = appendID(result.Duplicate, state.normalized.PathID)
			continue
		}
		if state.malformed {
			result.Malformed = appendID(result.Malformed, state.normalized.PathID)
			continue
		}
		outcome := evaluateRequirement(state.normalized, edges)
		switch outcome {
		case issueDuplicate:
			result.Duplicate = appendID(result.Duplicate, state.normalized.PathID)
		case issueMalformed:
			result.Malformed = appendID(result.Malformed, state.normalized.PathID)
		case issueMissingEvidence, issueMissingSnapshot:
			result.Missing = appendID(result.Missing, state.normalized.PathID)
		default:
			result.Complete = appendID(result.Complete, state.normalized.PathID)
		}
	}
	result.Duplicate = appendIDs(result.Duplicate, duplicateEdges)
	sortResultIDs(&result)
	result.Numerator = len(result.Complete)
	result.Status, result.Code = decision(result)
	return result
}

func prepareRequirements(requirements []Requirement) ([]requirementState, []semantic.ID) {
	states := make([]requirementState, 0, len(requirements))
	required := make([]semantic.ID, 0, len(requirements))
	seen := make(map[semantic.ID]struct{}, len(requirements))
	for _, raw := range requirements {
		state := requirementState{raw: raw, normalized: raw}
		if pathID, err := semantic.ParseIdentity(raw.PathID.String()); err == nil {
			state.normalized.PathID = pathID
		}
		normalized, err := normalizeRequirement(raw)
		if err != nil {
			state.malformed = true
		} else {
			state.normalized = normalized
		}
		if hasDuplicateRecordIDs(raw.RecordIDs) {
			state.duplicate = true
		}
		if state.normalized.PathID != "" {
			if _, exists := seen[state.normalized.PathID]; exists {
				state.duplicate = true
			} else {
				seen[state.normalized.PathID] = struct{}{}
			}
		}
		required = appendID(required, state.normalized.PathID)
		states = append(states, state)
	}
	sortIDs(required)
	return states, required
}

func hasDuplicateRecordIDs(ids []semantic.ID) bool {
	seen := make(map[semantic.ID]struct{}, len(ids))
	for _, rawID := range ids {
		id := rawID
		if normalized, err := semantic.ParseIdentity(rawID.String()); err == nil {
			id = normalized
		}
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}

func normalizeRequirement(raw Requirement) (Requirement, error) {
	pathID, err := semantic.ParseIdentity(raw.PathID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("path ID: %w", err)
	}
	start, err := semantic.ParseIdentity(raw.StartID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("start ID: %w", err)
	}
	end, err := semantic.ParseIdentity(raw.EndID.String())
	if err != nil {
		return Requirement{}, fmt.Errorf("end ID: %w", err)
	}
	if len(raw.RecordIDs) == 0 || len(raw.RecordIDs) != len(raw.ExpectedKinds) {
		return Requirement{}, errors.New("record and kind sequences must be non-empty and equal")
	}
	records := make([]semantic.ID, len(raw.RecordIDs))
	seen := make(map[semantic.ID]struct{}, len(records))
	for i, rawID := range raw.RecordIDs {
		recordID, parseErr := semantic.ParseIdentity(rawID.String())
		if parseErr != nil {
			return Requirement{}, fmt.Errorf("record ID: %w", parseErr)
		}
		if _, exists := seen[recordID]; exists {
			return Requirement{}, fmt.Errorf("duplicate record ID %s", recordID)
		}
		seen[recordID] = struct{}{}
		records[i] = recordID
		if !raw.ExpectedKinds[i].Valid() {
			return Requirement{}, fmt.Errorf("unknown inference kind %q", raw.ExpectedKinds[i])
		}
	}
	return Requirement{
		PathID: pathID, RecordIDs: records,
		ExpectedKinds: append([]semantic.InferenceKind(nil), raw.ExpectedKinds...),
		StartID:       start, EndID: end,
	}, nil
}
