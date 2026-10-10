package outcomedelta

import (
	"fmt"
	"slices"
)

// Compare reports supplied observations. It neither executes the program nor
// verifies the authenticity of its producer. A changed requirement is kept
// separate from regression under an unchanged requirement.
func Compare(before, after []byte) (*OutcomeDelta, error) {
	if digest(declaration) != "sha256:"+DeclarationSHA256 {
		return nil, fmt.Errorf("outcome declaration differs from generated structure")
	}
	b, err := runtimeInput(before)
	if err != nil {
		return nil, fmt.Errorf("before: %w", err)
	}
	a, err := runtimeInput(after)
	if err != nil {
		return nil, fmt.Errorf("after: %w", err)
	}
	bg, err := groups(b)
	if err != nil {
		return nil, fmt.Errorf("before: %w", err)
	}
	ag, err := groups(a)
	if err != nil {
		return nil, fmt.Errorf("after: %w", err)
	}
	r := &OutcomeDelta{Schema: "gooo/body-outcomes-delta/v1", DeclarationSha256: "sha256:" + DeclarationSHA256,
		BeforeInputSha256: digest(before), AfterInputSha256: digest(after), BeforeRuntime: runtimeSummary(b), AfterRuntime: runtimeSummary(a),
		Rows: []OutcomeChange{}, Counts: map[string]int{}, ComparatorOperations: map[string]int{"model_calls": 0, "program_executions": 0, "external_calls": 0, "repository_writes": 0},
		Scope: "Observed activity outcomes joined by the complete caller input tuple and stable activity ID; source contracts may change. Duplicate conflicting observations stay ambiguous. Supplied records are not execution attestation, unseen-input coverage, or causal performance evidence."}
	keys := make([]string, 0, len(bg)+len(ag))
	for k := range bg {
		keys = append(keys, k)
	}
	for k := range ag {
		if bg[k] == nil {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	for _, key := range keys {
		row := compareGroup(bg[key], ag[key])
		r.Rows = append(r.Rows, row)
		r.Counts["presence:"+row.Presence]++
		r.Counts["outcome:"+row.OutcomeChange]++
		r.Counts["requirement:"+row.RequirementChange]++
		r.Counts["assessment:"+row.Assessment]++
	}
	return r, nil
}

func records(g *group) []map[string]any {
	r := []map[string]any{}
	if g != nil {
		for _, o := range g.observations {
			r = append(r, o.record)
		}
	}
	return r
}

func ambiguous(g *group) bool {
	for _, o := range g.observations[1:] {
		if o.identity != g.observations[0].identity {
			return true
		}
	}
	return false
}

func compareGroup(b, a *group) OutcomeChange {
	g := b
	if g == nil {
		g = a
	}
	r := OutcomeChange{ActivityID: g.activity, RootInputs: g.inputs, Before: records(b), After: records(a),
		Presence: "BOTH", OutcomeChange: "UNOBSERVED", RequirementChange: "UNOBSERVED", Assessment: "UNOBSERVED"}
	if b == nil {
		r.Presence = "AFTER_ONLY"
		return r
	}
	if a == nil {
		r.Presence = "BEFORE_ONLY"
		return r
	}
	if ambiguous(b) || ambiguous(a) {
		r.OutcomeChange = "AMBIGUOUS"
		r.RequirementChange = "AMBIGUOUS"
		r.Assessment = "AMBIGUOUS"
		return r
	}
	before, after := b.observations[0], a.observations[0]
	if before.state != "UNOBSERVED" && after.state != "UNOBSERVED" {
		r.OutcomeChange = "UNCHANGED"
		if before.state != after.state || before.outcome != after.outcome {
			r.OutcomeChange = "CHANGED"
		}
	}
	if before.expected == "" || after.expected == "" {
		return r
	}
	r.RequirementChange = "UNCHANGED"
	if before.expected != after.expected {
		r.RequirementChange = "CHANGED"
		r.Assessment = "REQUIREMENT_CHANGED"
		return r
	}
	if before.passed == nil || after.passed == nil {
		return r
	}
	switch {
	case *before.passed && !*after.passed:
		r.Assessment = "REGRESSION"
	case !*before.passed && *after.passed:
		r.Assessment = "IMPROVEMENT"
	case *before.passed:
		r.Assessment = "STILL_MATCHING"
	default:
		r.Assessment = "STILL_FAILING"
	}
	return r
}
