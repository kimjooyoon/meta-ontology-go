package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SourceFillCandidateSet describes complete assignments in source order. The
// local chooser and a caller-guided search own separate selection observations.
type SourceFillCandidateSet struct {
	PlanSHA256 string                                `json:"plan_sha256"`
	Candidates []IRBodyFillCandidate                 `json:"candidates"`
	Generation *IRBodyFillCandidateGenerationReceipt `json:"generation,omitempty"`
}

// FillCandidate preserves both local suites. Only TestCasesPassed/Total may
// contribute to caller-guided selection; holdout observations stay separate.
type FillCandidate struct {
	Rejection            *IRBodyFillCandidateRejection         `json:"rejection,omitempty"`
	Schema               string                                `json:"schema"`
	Activity             string                                `json:"activity"`
	ActivityID           string                                `json:"activity_id"`
	InputSourceSHA256    string                                `json:"input_source_sha256"`
	SelectedSourceSHA256 string                                `json:"selected_source_sha256"`
	PlanSHA256           string                                `json:"plan_sha256"`
	CandidateCount       int                                   `json:"candidate_count"`
	Generation           *IRBodyFillCandidateGenerationReceipt `json:"candidate_generation,omitempty"`
	CandidateID          string                                `json:"candidate_id"`
	SelectionMethod      string                                `json:"selection_method"`
	HoleFills            []IRBodyFillHoleFill                  `json:"hole_fills"`
	TestCasesPassed      int                                   `json:"test_cases_passed"`
	TestCasesTotal       int                                   `json:"test_cases_total"`
	CaseResults          []IRBodyFillCaseResult                `json:"case_results,omitempty"`
	ValueCaseResults     []RecordAssemblyCase                  `json:"value_case_results,omitempty"`
	HoldoutCasesPassed   int                                   `json:"holdout_cases_passed"`
	HoldoutCasesTotal    int                                   `json:"holdout_cases_total"`
	HoldoutCaseResults   []IRBodyFillCaseResult                `json:"holdout_case_results,omitempty"`
	ValueHoldoutResults  []RecordAssemblyCase                  `json:"value_holdout_results,omitempty"`
}

func sourceFillCandidatePlan(ctx context.Context, filename string, source []byte, activity string) (IRBodyFillPlan, SourceFillCandidateSet, error) {
	var set SourceFillCandidateSet
	if ctx == nil {
		return IRBodyFillPlan{}, set, fmt.Errorf("source fill candidate requires context")
	}
	if err := ctx.Err(); err != nil {
		return IRBodyFillPlan{}, set, err
	}
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return IRBodyFillPlan{}, set, err
	}
	if !IsSourceIRBodyFill(spec) {
		return IRBodyFillPlan{}, set, fmt.Errorf("source fill candidate requires a source_fill contract")
	}
	if err = spec.Validate(); err != nil {
		return IRBodyFillPlan{}, set, err
	}
	plan, grammar, err := sourceIRBodyFillPlan(filename, source, activity, spec)
	if err != nil {
		return plan, set, err
	}
	if err = validateIRBodyFillPlan(plan); err != nil {
		return plan, set, err
	}
	raw, err := json.Marshal(plan)
	set = SourceFillCandidateSet{PlanSHA256: digest(raw), Candidates: plan.Candidates, Generation: grammar}
	return plan, set, err
}

func PlanSourceFillCandidates(ctx context.Context, filename string, source []byte, activity string) (SourceFillCandidateSet, error) {
	_, set, err := sourceFillCandidatePlan(ctx, filename, source, activity)
	return set, err
}

// RealizeSourceFillCandidate checks a complete source-owned assignment, including
// a locally imperfect one. It loads no model and never replaces a local winner's
// receipt or expectations. Deterministic type/training failures return a bound
// rejection; malformed source, cancellation and holdout failures remain terminal.
func RealizeSourceFillCandidate(ctx context.Context, filename string, source []byte, activity, candidateID string) (FillCandidate, []byte, error) {
	plan, set, err := sourceFillCandidatePlan(ctx, filename, source, activity)
	if err != nil {
		return FillCandidate{}, nil, err
	}
	candidate, found := candidateByID(plan.Candidates, candidateID)
	if !found {
		return FillCandidate{}, nil, fmt.Errorf("fill candidate is absent from the source assignments")
	}
	selected, err := selectedFillSource(filename, source, activity, plan, candidate)
	if err != nil {
		return FillCandidate{}, nil, err
	}
	id, err := fillActivityID(filename, source, activity)
	if err != nil {
		return FillCandidate{}, nil, err
	}
	observation := FillCandidate{Schema: "gooo/fill-candidate/v1", Activity: activity, ActivityID: id,
		InputSourceSHA256: digest(source), SelectedSourceSHA256: digest(selected), PlanSHA256: set.PlanSHA256,
		CandidateCount: len(set.Candidates), Generation: set.Generation, CandidateID: candidate.ID,
		SelectionMethod: "caller_selected_assignment", HoleFills: bodyFillHoleResults(bodyFillPlanHoles(plan), candidate.Fills)}
	generated, err := GenerateWithPlanner(ctx, filename, selected, activity, "", "")
	if err != nil {
		err = rejectSourceFillCandidate(ctx, &observation, "TYPECHECK", err)
		return observation, nil, err
	}
	if err = scoreSourceFillCandidate(ctx, filename, selected, generated.Source, activity, plan, &observation); err != nil {
		return observation, nil, err
	}
	return observation, selected, ctx.Err()
}

func fillActivityID(filename string, source []byte, activity string) (string, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return "", fmt.Errorf("parse source fill identity: %v", diagnostics)
	}
	model, _, err := resolveBodyModel(file)
	if err != nil {
		return "", err
	}
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activity {
			return string(node.ID), nil
		}
	}
	return "", fmt.Errorf("activity %q has no stable semantic identity", activity)
}

func rejectSourceFillCandidate(ctx context.Context, r *FillCandidate, stage string, err error) error {
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	r.Rejection = &IRBodyFillCandidateRejection{CandidateID: r.CandidateID, Stage: stage, Reason: err.Error(), HoleFills: r.HoleFills}
	return &SourceFillCandidateRejection{Observation: *r.Rejection}
}

func selectedFillSource(filename string, source []byte, activity string, plan IRBodyFillPlan, candidate IRBodyFillCandidate) ([]byte, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return nil, fmt.Errorf("parse source fill candidate: %v", diagnostics)
	}
	for _, declaration := range file.Declarations {
		body, ok := declaration.(*syntax.ActivityDecl)
		if !ok || body.Name != activity {
			continue
		}
		filled := body.ValueProgram
		for _, hole := range bodyFillPlanHoles(plan) {
			var err error
			filled, err = replaceIdentifier(filled, bodyFillHoleToken(hole.ID), candidate.Fills[hole.ID])
			if err != nil {
				return nil, err
			}
		}
		selected, err := replaceActivityProgram(source, body.ValueProgramSpan, filled)
		if err != nil {
			return nil, err
		}
		return sourceWithoutIRBodyFill(filename, selected, activity)
	}
	return nil, fmt.Errorf("source fill activity %q was not found", activity)
}
