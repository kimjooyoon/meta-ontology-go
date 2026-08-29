package metrictransition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func validateCausalUnknowns(ledger transformationeffect.Ledger, report generation.ReceiptReport) error {
	projection, err := deriveCausalUnknowns(report)
	if err != nil {
		return err
	}
	if ledger.DirectUnknownCount != projection.DirectUnknownCount ||
		ledger.DependencyBlockedUnknownCount != projection.DependencyBlockedUnknownCount ||
		ledger.UnknownCausalDigest != projection.Digest {
		return fmt.Errorf("metric transition causal unknown projection diverged")
	}
	if ledger.OperationOutcome == effectOutcomeMixedRefuted &&
		(projection.DirectUnknownCount != 0 || projection.DependencyBlockedUnknownCount != len(report.Unknowns)) {
		return fmt.Errorf("metric transition mixed outcome has unbound unknowns")
	}
	return nil
}

func deriveCausalUnknowns(report generation.ReceiptReport) (transformationeffect.CausalUnknownProjection, error) {
	projection := transformationeffect.CausalUnknownProjection{Records: []transformationeffect.CausalUnknownRecord{}}
	failures := make(map[string]generation.ObservationFailure, len(report.Failures))
	for _, failure := range report.Failures {
		if failure.ActionIndicatorID == "" || failure.Decision != string(generation.ReceiptDecisionRefuted) ||
			failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			failure.NextOperation == "" || failure.BlockedBy == nil {
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("causal failure context is incomplete")
		}
		if _, exists := failures[failure.ActionIndicatorID]; exists {
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("causal failure context is duplicated")
		}
		failures[failure.ActionIndicatorID] = failure
	}
	seen := make(map[string]bool, len(report.Unknowns))
	for _, unknown := range report.Unknowns {
		if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" || unknown.Operation == "" ||
			unknown.Activity == "" || unknown.Output == "" || unknown.Executor == "" || unknown.Evaluator == "" ||
			unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.NextOperation == "" ||
			unknown.BlockedBy == nil {
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("causal unknown context is incomplete")
		}
		key := unknown.ActionIndicatorID + "\x00" + unknown.RequiredIndicatorID
		if seen[key] {
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("causal unknown obligation is duplicated")
		}
		seen[key] = true
		failure, hasFailure := failures[unknown.ActionIndicatorID]
		switch unknown.UnknownClass {
		case generation.ReceiptUnknownClassDirectMissing,
			generation.ReceiptUnknownClassMalformedEvidence,
			generation.ReceiptUnknownClassUnexpectedEvidence:
			if len(unknown.BlockedBy) != 0 {
				return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("direct unknown has a frontier")
			}
			projection.DirectUnknownCount++
		case generation.ReceiptUnknownClassDependencyBlocked:
			root := "operation-failure:" + unknown.ActionIndicatorID
			if !hasFailure || failure.Stage != unknown.Stage || failure.Step != unknown.Step ||
				failure.Reason != string(unknown.Reason) || failure.NextOperation != unknown.NextOperation ||
				len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != root {
				return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("dependency unknown is not bound")
			}
			projection.DependencyBlockedUnknownCount++
		default:
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("causal unknown class is unsupported")
		}
		projection.Records = append(projection.Records, transformationeffect.CausalUnknownRecord{
			ActionIndicatorID: unknown.ActionIndicatorID, RequiredIndicatorID: unknown.RequiredIndicatorID,
			Stage: unknown.Stage, Step: unknown.Step, Reason: string(unknown.Reason),
			UnknownClass: unknown.UnknownClass, NextOperation: unknown.NextOperation,
			BlockedBy: append([]string{}, unknown.BlockedBy...)})
	}
	sort.Slice(projection.Records, func(left, right int) bool {
		return metricCausalUnknownKey(projection.Records[left]) < metricCausalUnknownKey(projection.Records[right])
	})
	payload, err := json.Marshal(projection)
	if err != nil {
		return transformationeffect.CausalUnknownProjection{}, err
	}
	digest := sha256.Sum256(payload)
	projection.Digest = hex.EncodeToString(digest[:])
	return projection, nil
}

func metricCausalUnknownKey(record transformationeffect.CausalUnknownRecord) string {
	return strings.Join([]string{record.ActionIndicatorID, record.RequiredIndicatorID,
		record.Stage, record.Step, record.Reason, record.UnknownClass, record.NextOperation,
		strings.Join(record.BlockedBy, "\x00")}, "\x00")
}
