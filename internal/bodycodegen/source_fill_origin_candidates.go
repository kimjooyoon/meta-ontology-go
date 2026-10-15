package bodycodegen

import "context"

type SourceFillOriginCandidate struct {
	CandidateID string               `json:"candidate_id"`
	HoleFills   []IRBodyFillHoleFill `json:"hole_fills"`
}

// SourceFillOriginCandidates is the source-bounded list a caller can choose
// from before inspecting origins. Generation coverage is separate from accuracy.
type SourceFillOriginCandidates struct {
	Schema       string                                `json:"schema"`
	Filename     string                                `json:"filename"`
	SourceSHA256 string                                `json:"source_sha256"`
	Activity     string                                `json:"activity"`
	ActivityID   string                                `json:"activity_id"`
	Kind         string                                `json:"kind"`
	PlanSHA256   string                                `json:"plan_sha256"`
	Candidates   []SourceFillOriginCandidate           `json:"candidates"`
	Generation   *IRBodyFillCandidateGenerationReceipt `json:"candidate_generation,omitempty"`
	Scope        string                                `json:"scope"`
}

func InspectSourceFillOriginCandidates(ctx context.Context, filename string, source []byte, activity string) (SourceFillOriginCandidates, error) {
	r := SourceFillOriginCandidates{Schema: "gooo/source-fill-origin-candidates/v1", Filename: filename,
		SourceSHA256: digest(source), Activity: activity, Kind: "declared_assignments",
		Scope: "source-owned finite assignments; grammar coverage only; no candidate selection, body materialization, case evaluation or model call"}
	plan, set, err := sourceFillCandidatePlan(ctx, filename, source, activity)
	if err != nil {
		return r, err
	}
	r.ActivityID, err = fillActivityID(filename, source, activity)
	if err != nil {
		return r, err
	}
	r.PlanSHA256, r.Generation = set.PlanSHA256, set.Generation
	if set.Generation != nil {
		r.Kind = "derived_grammar_assignments"
	}
	for _, candidate := range set.Candidates {
		r.Candidates = append(r.Candidates, SourceFillOriginCandidate{CandidateID: candidate.ID,
			HoleFills: bodyFillHoleResults(bodyFillPlanHoles(plan), bodyFillCandidateFills(plan, candidate))})
	}
	return r, ctx.Err()
}

func IsSourceFillDerivation(ctx context.Context, filename string, source []byte, activity string) (bool, error) {
	spec, err := SourceAssembly(ctx, filename, source, activity)
	return IsSourceIRBodyFill(spec) && spec.FillPlan.Generation != nil, err
}
