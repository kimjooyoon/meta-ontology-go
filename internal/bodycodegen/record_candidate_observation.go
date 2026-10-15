package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type RecordCandidateObservation struct {
	ReceiptSHA256        string               `json:"receipt_sha256"`
	OriginalSourceSHA256 string               `json:"original_source_sha256"`
	EffectiveCasesSHA256 string               `json:"effective_cases_sha256"`
	InitialCases         int                  `json:"initial_cases"`
	AddedCases           int                  `json:"added_cases"`
	OracleActivityID     string               `json:"oracle_activity_id,omitempty"`
	OracleGoSHA256       string               `json:"oracle_go_sha256,omitempty"`
	OracleCases          []RecordAssemblyCase `json:"oracle_cases,omitempty"`
}

// Retain the already acquired labels. Caller construction verifies its initial
// generation separately; this step binds those labels to the current source and
// rechecks the fixed oracle without asking or ranking any additional question.
func retainRecordCandidateObservation(ctx context.Context, filename string, source []byte, p *recordAssemblyPlan,
	prior *RecordAssemblyReceipt) (*RecordCandidateObservation, error) {
	if isCriteriaQuestionPlan(*p) {
		return nil, nil
	}
	if p.spec.Observation == nil {
		if prior != nil && prior.Observation != nil {
			return nil, fmt.Errorf("record observation declaration was removed")
		}
		return nil, nil
	}
	if prior == nil || prior.Observation == nil {
		return nil, fmt.Errorf("record candidate requires its retained source observation")
	}
	r, err := canonicalRecordObservation(prior.Observation)
	if err != nil {
		return nil, err
	}
	cases, err := retainedRecordObservationCases(p, prior, r)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(r)
	bound := &RecordCandidateObservation{ReceiptSHA256: digest(raw), OriginalSourceSHA256: r.SourceSHA256,
		EffectiveCasesSHA256: recordCasesSHA(cases), InitialCases: len(p.spec.ValueCases), AddedCases: len(cases) - len(p.spec.ValueCases)}
	check := &RecordObservationReceipt{Options: r.Options}
	_, err = prepareRecordOracle(ctx, filename, source, *p, cases, check)
	bound.OracleActivityID, bound.OracleGoSHA256, bound.OracleCases = check.OracleActivityID, check.OracleGoSHA256, check.OracleInitialCases
	if err != nil {
		return bound, fmt.Errorf("record candidate changed retained answers: %w", err)
	}
	if bound.OracleActivityID != r.OracleActivityID {
		return bound, fmt.Errorf("record candidate oracle identity differs")
	}
	p.effectiveCases = cases
	return bound, nil
}

func retainedRecordObservationCases(p *recordAssemblyPlan, prior *RecordAssemblyReceipt,
	r *RecordObservationReceipt) ([]assemblyspec.ValueCase, error) {
	contract, _ := p.spec.Canonical()
	options, _ := json.Marshal(p.spec.Observation)
	origin := prior.OriginalSourceSHA256
	if prior.RankingSourceSHA256 != "" {
		origin = prior.RankingSourceSHA256
	}
	if r.Schema != "gooo/record-observation/v1" || r.Status == "FAILED" || r.Status == "STARTED" ||
		r.SourceSHA256 != origin || r.ContractSHA256 != digest([]byte(contract)) || prior.ContractSHA256 != r.ContractSHA256 ||
		r.OptionsSHA256 != digest(options) || !reflect.DeepEqual(r.Options, *p.spec.Observation) ||
		!reflect.DeepEqual(r.InitialCases, p.spec.ValueCases) || len(r.Rounds) < 1 || len(r.Rounds) > r.Options.MaxRounds+1 {
		return nil, fmt.Errorf("record candidate observation differs from its source contract")
	}
	cases := append([]assemblyspec.ValueCase(nil), p.spec.ValueCases...)
	seen := make(map[string]bool)
	for _, c := range cases {
		key, _ := assemblyspec.RecordValueKey(c.Inputs)
		seen[key] = true
	}
	for i, round := range r.Rounds {
		c := round.Observation
		if c == nil {
			continue
		}
		if r.Options.OracleActivity == "" || i >= r.Options.MaxRounds ||
			round.RecommendedIndex == nil || *round.RecommendedIndex < 0 || *round.RecommendedIndex >= len(round.Probes) ||
			round.Probes[*round.RecommendedIndex].Input != c.Inputs || !slices.Contains(r.Options.ValueInputs, c.Inputs) {
			return nil, fmt.Errorf("record acquired answer has no declared bounded question")
		}
		key, err := assemblyspec.RecordValueKey(c.Inputs)
		if err != nil || seen[key] {
			return nil, fmt.Errorf("record acquired input is invalid or repeated")
		}
		canonical, err := assemblyspec.CanonicalValue(c.Expected)
		if err != nil || canonical != c.Expected {
			return nil, fmt.Errorf("record acquired answer is not canonical")
		}
		seen[key] = true
		cases = append(cases, *c)
	}
	criteriaInputs := 0
	if p.spec.Criteria != nil {
		criteriaInputs = len(p.spec.Criteria.Inputs) + len(p.spec.Criteria.HoldoutInputs)
	}
	if len(cases)+len(p.spec.ValueHoldoutCases)+criteriaInputs > 128 || r.EffectiveCases != len(cases) || r.EffectiveCasesSHA256 != recordCasesSHA(cases) {
		return nil, fmt.Errorf("record effective cases differ from retained observation")
	}
	return cases, nil
}
