package transformationeffect

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

// CausalUnknownRecord is the stable, receipt-derived identity of one unknown
// obligation.  It deliberately retains the failure context that caused a
// dependency-blocked obligation.
type CausalUnknownRecord struct {
	ActionIndicatorID   string   `json:"action_indicator_id"`
	RequiredIndicatorID string   `json:"required_indicator_id"`
	Stage               string   `json:"stage"`
	Step                string   `json:"step"`
	Reason              string   `json:"reason"`
	UnknownClass        string   `json:"unknown_class"`
	NextOperation       string   `json:"next_operation"`
	BlockedBy           []string `json:"blocked_by"`
}

// CausalUnknownProjection is the canonical aggregate consumed by downstream
// verifiers. Digest is excluded from its own input by design.
type CausalUnknownProjection struct {
	DirectUnknownCount           int                   `json:"direct_unknown_count"`
	DependencyBlockedUnknownCount int                  `json:"dependency_blocked_unknown_count"`
	Records                      []CausalUnknownRecord `json:"records"`
	Digest                       string                `json:"-"`
}

func BuildCausalUnknownProjection(report generation.ReceiptReport) (CausalUnknownProjection, error) {
	projection := CausalUnknownProjection{Records: []CausalUnknownRecord{}}
	failures := make(map[string]generation.ObservationFailure, len(report.Failures))
	for _, failure := range report.Failures {
		if failure.ActionIndicatorID == "" || failure.Decision != string(generation.ReceiptDecisionRefuted) ||
			failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			failure.NextOperation == "" || failure.BlockedBy == nil {
			return CausalUnknownProjection{}, fmt.Errorf("causal unknown failure is incomplete")
		}
		if _, exists := failures[failure.ActionIndicatorID]; exists {
			return CausalUnknownProjection{}, fmt.Errorf("causal unknown failure is duplicated")
		}
		failures[failure.ActionIndicatorID] = failure
	}
	seen := make(map[string]bool, len(report.Unknowns))
	for _, unknown := range report.Unknowns {
		if err := validateCausalUnknown(unknown); err != nil {
			return CausalUnknownProjection{}, err
		}
		key := unknown.ActionIndicatorID + "\x00" + unknown.RequiredIndicatorID
		if seen[key] {
			return CausalUnknownProjection{}, fmt.Errorf("causal unknown obligation is duplicated")
		}
		seen[key] = true
		failure, hasFailure := failures[unknown.ActionIndicatorID]
		switch unknown.UnknownClass {
		case generation.ReceiptUnknownClassDirectMissing,
			generation.ReceiptUnknownClassMalformedEvidence,
			generation.ReceiptUnknownClassUnexpectedEvidence:
			if len(unknown.BlockedBy) != 0 {
				return CausalUnknownProjection{}, fmt.Errorf("direct unknown has a dependency frontier")
			}
			projection.DirectUnknownCount++
		case generation.ReceiptUnknownClassDependencyBlocked:
			root := "operation-failure:" + unknown.ActionIndicatorID
			if !hasFailure || failure.Decision != string(generation.ReceiptDecisionRefuted) ||
				failure.Stage != unknown.Stage || failure.Step != unknown.Step ||
				failure.Reason != string(unknown.Reason) || failure.NextOperation != unknown.NextOperation ||
				len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != root {
				return CausalUnknownProjection{}, fmt.Errorf("dependency unknown frontier is not bound")
			}
			projection.DependencyBlockedUnknownCount++
		default:
			return CausalUnknownProjection{}, fmt.Errorf("causal unknown class is unsupported")
		}
		projection.Records = append(projection.Records, CausalUnknownRecord{
			ActionIndicatorID: unknown.ActionIndicatorID, RequiredIndicatorID: unknown.RequiredIndicatorID,
			Stage: unknown.Stage, Step: unknown.Step, Reason: string(unknown.Reason),
			UnknownClass: unknown.UnknownClass, NextOperation: unknown.NextOperation,
			BlockedBy: append([]string{}, unknown.BlockedBy...)})
	}
	sort.Slice(projection.Records, func(left, right int) bool {
		return causalUnknownKey(projection.Records[left]) < causalUnknownKey(projection.Records[right])
	})
	projection.Digest = hashJSON(projection)
	return projection, nil
}

func validateCausalUnknown(unknown generation.ReceiptUnknown) error {
	if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" || unknown.Operation == "" ||
		unknown.Activity == "" || unknown.Output == "" || unknown.Executor == "" || unknown.Evaluator == "" ||
		unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.NextOperation == "" ||
		unknown.BlockedBy == nil {
		return fmt.Errorf("causal unknown is missing required fields")
	}
	return nil
}

func causalUnknownKey(record CausalUnknownRecord) string {
	return strings.Join([]string{record.ActionIndicatorID, record.RequiredIndicatorID,
		record.Stage, record.Step, record.Reason, record.UnknownClass, record.NextOperation,
		strings.Join(record.BlockedBy, "\x00")}, "\x00")
}
