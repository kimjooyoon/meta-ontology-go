package bodycodegen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type CriteriaQuestionRetention struct {
	ReceiptSHA256         string   `json:"receipt_sha256"`
	SourceSHA256          string   `json:"source_sha256"`
	EffectiveInputsSHA256 string   `json:"effective_inputs_sha256"`
	InitialInputs         int      `json:"initial_inputs"`
	AcquiredInputs        []string `json:"acquired_inputs"`
}

// Caller construction verifies the original receipt before changing callees.
// Retention binds its input obligations without asking the questions again.
func retainCriteriaQuestions(p *recordAssemblyPlan, prior *RecordAssemblyReceipt) (*CriteriaQuestionRetention, error) {
	if !isCriteriaQuestionPlan(*p) {
		if prior != nil && prior.CriteriaQuestions != nil {
			return nil, fmt.Errorf("criteria questions were removed from the source")
		}
		return nil, nil
	}
	if prior == nil || prior.CriteriaQuestions == nil {
		return nil, fmt.Errorf("record candidate requires its retained criteria question history")
	}
	r := prior.CriteriaQuestions
	contract, _ := p.spec.Canonical()
	origin := prior.OriginalSourceSHA256
	if prior.RankingSourceSHA256 != "" {
		origin = prior.RankingSourceSHA256
	}
	if prior.Observation != nil || r.Schema != criteriaQuestionSchema(r.Options) || r.Status == "STARTED" || r.Status == "FAILED" ||
		r.SourceSHA256 != origin || r.ContractSHA256 != digest([]byte(contract)) || prior.ContractSHA256 != r.ContractSHA256 ||
		!reflect.DeepEqual(r.Options, *p.spec.Observation) || !slices.Equal(r.InitialInputs, p.spec.Criteria.Inputs) ||
		len(r.Rounds) < 1 || len(r.Rounds) > r.Options.MaxRounds+1 {
		return nil, fmt.Errorf("criteria question history differs from the source contract")
	}
	inputs := append([]string(nil), p.spec.Criteria.Inputs...)
	seen := map[string]bool{}
	for _, input := range inputs {
		key, _ := assemblyspec.RecordValueKey(input)
		seen[key] = true
	}
	var acquired []string
	for i, round := range r.Rounds {
		input := round.AcquiredInput
		if input == "" {
			continue
		}
		if i >= r.Options.MaxRounds || round.RecommendedIndex == nil || *round.RecommendedIndex < 0 || *round.RecommendedIndex >= len(round.Probes) ||
			round.Probes[*round.RecommendedIndex].Input != input || !validCriteriaQuestionReason(r.Options, round) || !slices.Contains(r.Options.CriteriaInputs, input) {
			return nil, fmt.Errorf("acquired criteria input has no declared question reason")
		}
		key, err := assemblyspec.RecordValueKey(input)
		if err != nil || seen[key] {
			return nil, fmt.Errorf("acquired criteria input is invalid or repeated")
		}
		seen[key] = true
		inputs = append(inputs, input)
		acquired = append(acquired, input)
	}
	if !slices.Equal(acquired, r.AcquiredInputs) || len(inputs)+len(p.spec.Criteria.HoldoutInputs)+len(p.spec.ValueCases)+len(p.spec.ValueHoldoutCases) > 128 ||
		r.EffectiveInputs != len(inputs) || r.EffectiveInputsSHA256 != criteriaInputsSHA(inputs) {
		return nil, fmt.Errorf("effective criteria inputs differ from history")
	}
	p.effectiveCriteriaInputs = inputs
	raw, _ := json.Marshal(r)
	return &CriteriaQuestionRetention{ReceiptSHA256: digest(raw), SourceSHA256: r.SourceSHA256, EffectiveInputsSHA256: r.EffectiveInputsSHA256,
		InitialInputs: len(r.InitialInputs), AcquiredInputs: append([]string(nil), acquired...)}, nil
}
