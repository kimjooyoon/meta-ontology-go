package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// The original observation remains immutable. These fields bind its acquired
// obligations to the candidate's current source, including earlier slot edits.
type PathCandidateObservation struct {
	ReceiptSHA256        string                 `json:"receipt_sha256"`
	OriginalSourceSHA256 string                 `json:"original_source_sha256"`
	EffectiveCasesSHA256 string                 `json:"effective_cases_sha256"`
	InitialCases         int                    `json:"initial_cases"`
	AddedCases           int                    `json:"added_cases"`
	OracleActivityID     string                 `json:"oracle_activity_id,omitempty"`
	OracleGoSHA256       string                 `json:"oracle_go_sha256,omitempty"`
	OracleCases          []IRBodyFillCaseResult `json:"oracle_cases,omitempty"`
}

// Carry acquired answers forward without ranking candidates or asking new
// questions. A changed oracle must still satisfy the retained finite answers.
// Candidate checking cannot relabel them to accommodate another source choice.
func sourcePathCandidateCases(ctx context.Context, filename string, source []byte, activity string,
	document pathplan.Document, prior *BodyPathReceipt, result *PathCandidate) ([]pathplan.TestCase, error) {
	result.Stage = "OBSERVED_CASES"
	if err := verifySourcePathObservation(ctx, filename, source, activity, prior.Observation); err != nil {
		return nil, err
	}
	r := prior.Observation
	if r == nil {
		return document.TestCases, nil
	}
	sha, _, bad := observedPathContract(prior)
	if bad || r.Status == "FAILED" || r.PreparedPlanSHA256 != result.PlanSHA256 ||
		!reflect.DeepEqual(r.InitialCases, document.TestCases) {
		return nil, fmt.Errorf("typed path candidate observation differs from its source contract")
	}
	cases := append([]pathplan.TestCase(nil), document.TestCases...)
	for _, round := range r.Rounds {
		if round.Observation != nil {
			cases = append(cases, *round.Observation)
		}
	}
	raw, _ := json.Marshal(r)
	bound := &PathCandidateObservation{ReceiptSHA256: digest(raw), OriginalSourceSHA256: r.SourceSHA256,
		EffectiveCasesSHA256: sha, InitialCases: len(document.TestCases), AddedCases: len(cases) - len(document.TestCases)}
	result.Schema, result.Observation = "gooo/path-candidate/v2", bound
	current := &PathObservationReceipt{Options: r.Options}
	_, err := preparePathOracle(ctx, filename, source, activity, cases, current)
	bound.OracleActivityID, bound.OracleGoSHA256 = current.OracleActivityID, current.OracleGoSHA256
	bound.OracleCases = current.OracleInitialCases
	if err != nil {
		return nil, fmt.Errorf("retained observation no longer holds in candidate source: %w", err)
	}
	if bound.OracleActivityID != r.OracleActivityID {
		return nil, fmt.Errorf("candidate oracle activity identity differs from the retained observation")
	}
	return cases, nil
}
