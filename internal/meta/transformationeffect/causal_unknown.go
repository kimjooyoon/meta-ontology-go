package transformationeffect

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

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
