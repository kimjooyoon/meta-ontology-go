package completenessdelta

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// Compare consumes immutable receipt observations. A successful return means
// the comparison was produced, not that either product passed its obligations.
func Compare(beforeInput, afterInput []byte) (*CompletenessDelta, error) {
	if !declarationMatches() {
		return nil, fmt.Errorf("delta declaration differs from generated structure")
	}
	beforeBytes, err := receiptBytes(beforeInput)
	if err != nil {
		return nil, fmt.Errorf("before: %w", err)
	}
	afterBytes, err := receiptBytes(afterInput)
	if err != nil {
		return nil, fmt.Errorf("after: %w", err)
	}
	before, err := completeness.Decode(beforeBytes)
	if err != nil {
		return nil, err
	}
	after, err := completeness.Decode(afterBytes)
	if err != nil {
		return nil, err
	}
	relation, basis := scopeRelation(before, after, digest(beforeBytes))
	r := &CompletenessDelta{
		Schema: CompletenessDeltaSchema, Status: "PROGRESS", Relation: relation, Basis: basis,
		Declaration:       map[string]any{"declaration_sha256": DeclarationSHA256, "projection_profile": ProjectionProfile},
		BeforeInputSha256: digest(beforeInput), AfterInputSha256: digest(afterInput),
		BeforeReceiptSha256: digest(beforeBytes), AfterReceiptSha256: digest(afterBytes),
		BeforeReceipt: record(beforeBytes), AfterReceipt: record(afterBytes),
		ChangedScopeFields: changedScope(before.Scope, after.Scope),
		Dimensions:         []DimensionChange{}, TransitionCounts: map[string]int{},
		ComparatorOperations: map[string]int{"model_calls": 0, "external_provider_calls": 0, "generated_program_executions": 0, "repository_writes": 0},
		NotClaimed:           []string{"aggregate intent completeness", "execution or producer attestation", "causal model improvement", "permission grants", "unseen behavior or production workflow coverage"},
	}
	if relation == "SAME_MEASUREMENT_SCOPE" {
		r.Status = "PASS"
	}
	appendChanges(r, before.Dimensions, after.Dimensions)
	return r, nil
}

func appendChanges(r *CompletenessDelta, before, after []completeness.CompletenessDimension) {
	afterByID := map[string]completeness.CompletenessDimension{}
	beforeIDs := map[string]bool{}
	for _, d := range after {
		afterByID[d.ID] = d
	}
	for _, b := range before {
		beforeIDs[b.ID] = true
		var a *DimensionObservation
		if found, ok := afterByID[b.ID]; ok {
			a = snapshot(found)
		}
		r.Dimensions = append(r.Dimensions, change(b.ID, snapshot(b), a, r.Relation))
	}
	for _, a := range after {
		if !beforeIDs[a.ID] {
			r.Dimensions = append(r.Dimensions, change(a.ID, nil, snapshot(a), r.Relation))
		}
	}
	for _, d := range r.Dimensions {
		r.TransitionCounts["presence:"+d.Presence]++
		r.TransitionCounts["state:"+d.StateTransition]++
		r.TransitionCounts["observation:"+d.ObservationChange]++
		r.TransitionCounts["count:"+d.CountDirection]++
		r.TransitionCounts["regression:"+d.Regression]++
	}
}

func snapshot(d completeness.CompletenessDimension) *DimensionObservation {
	return &DimensionObservation{Status: d.Status, Numerator: d.Numerator, Denominator: d.Denominator,
		Unit: d.Unit, Reason: d.Reason, Evidence: append([]string{}, d.Evidence...)}
}
