package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"math/bits"
	"slices"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type SourcePathCandidateSet struct {
	DocumentSHA256 string   `json:"document_sha256"`
	PlanSHA256     string   `json:"plan_sha256"`
	AttemptBudget  int      `json:"attempt_budget"`
	Declared       int      `json:"declared_combinations"`
	Masks          []uint16 `json:"masks"`
	Order          string   `json:"order"`
}

type PathCandidate struct {
	Schema               string                 `json:"schema"`
	Activity             string                 `json:"activity"`
	ActivityID           string                 `json:"activity_id"`
	InputSourceSHA256    string                 `json:"input_source_sha256"`
	SelectedSourceSHA256 string                 `json:"selected_source_sha256,omitempty"`
	DocumentSHA256       string                 `json:"document_sha256"`
	PlanSHA256           string                 `json:"plan_sha256"`
	Mask                 uint16                 `json:"mask"`
	Choices              map[string]string      `json:"choices"`
	CandidateSet         SourcePathCandidateSet `json:"candidate_set"`
	Stage                string                 `json:"stage"`
	Failure              string                 `json:"failure,omitempty"`
	Cases                []IRBodyFillCaseResult `json:"case_results,omitempty"`
	Passed               int                    `json:"local_passed"`
	Total                int                    `json:"local_total"`
}

type SourcePathCandidateRejection struct{ cause error }

func (e *SourcePathCandidateRejection) Error() string { return e.cause.Error() }
func (e *SourcePathCandidateRejection) Unwrap() error { return e.cause }

func IsSourceTypedPathAssembly(spec *assemblyspec.Spec) bool {
	return spec != nil && len(spec.Choices) != 0 && !IsRecordAssembly(spec) &&
		!IsSourceIRSearch(spec) && !IsSourceIRBodyFill(spec)
}

// PlanSourcePathCandidates retains the initial selected body as the first row.
// Remaining rows follow fallback distance, then numeric mask. That explicit
// deterministic order is separate from historical initial model ranking. No
// inference or finite candidate execution occurs here. The initial row consumes
// one of the source's 1..64 attempts, including when a model chose it.
func PlanSourcePathCandidates(ctx context.Context, filename string, source []byte, activity string,
	prior *BodyPathReceipt) (SourcePathCandidateSet, error) {
	var set SourcePathCandidateSet
	document, err := DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	if err != nil || prior == nil {
		return set, fmt.Errorf("typed path candidates require a source contract and initial receipt: %v", err)
	}
	prepared, err := document.Prepare()
	if err != nil {
		return set, err
	}
	raw, _ := json.Marshal(document)
	selectionSHA := bodyPathSelection(prior).PlanSHA256
	originalSHA := prepared.PlanSHA256()
	rankedSHA := originalSHA
	if context := prior.ModelContext; context != nil && context.Status == "ENCODED" {
		if context.OriginalPlanSHA != originalSHA || context.RankedPlanSHA == "" {
			return set, fmt.Errorf("typed path model context differs from its original source plan")
		}
		rankedSHA = context.RankedPlanSHA
	}
	if digest(raw) != prior.DocumentSHA256 || rankedSHA != selectionSHA {
		return set, fmt.Errorf("typed path candidate document differs from initial construction")
	}
	choices := bodyPathSelection(prior).Choices
	if _, err := prepared.Compile(choices); err != nil {
		return set, fmt.Errorf("initial typed path selection differs: %w", err)
	}
	var initial, fallback uint16
	for i, choice := range document.Plan.Decisions {
		if choices[choice.ID] == choice.Options[1].Label {
			initial |= 1 << i
		}
		if choice.Fallback == choice.Options[1].Label {
			fallback |= 1 << i
		}
	}
	set = SourcePathCandidateSet{DocumentSHA256: digest(raw), PlanSHA256: prepared.PlanSHA256(),
		AttemptBudget: document.MaxAttempts, Declared: 1 << len(document.Plan.Decisions),
		Masks: []uint16{initial}, Order: "initial_selection_then_fallback_distance_then_numeric_mask"}
	limit := min(set.AttemptBudget, set.Declared)
	for distance := 0; len(set.Masks) < limit && distance <= len(document.Plan.Decisions); distance++ {
		for mask := 0; len(set.Masks) < limit && mask < set.Declared; mask++ {
			if err := ctx.Err(); err != nil {
				return SourcePathCandidateSet{}, err
			}
			if uint16(mask) != initial && bits.OnesCount16(uint16(mask)^fallback) == distance {
				set.Masks = append(set.Masks, uint16(mask))
			}
		}
	}
	return set, nil
}

// RealizeSourcePathCandidate rebinds the complete current source before checking
// a combined typed selection. Local cases remain local. The caller must choose
// a mask from PlanSourcePathCandidates; this API rederives that same bound.
// Model receipts and original source are never rewritten by this observation.
func RealizeSourcePathCandidate(ctx context.Context, filename string, source []byte, activity string,
	prior *BodyPathReceipt, mask uint16) (PathCandidate, []byte, error) {
	var result PathCandidate
	set, err := PlanSourcePathCandidates(ctx, filename, source, activity, prior)
	if err != nil {
		return result, nil, err
	}
	if !slices.Contains(set.Masks, mask) {
		return result, nil, fmt.Errorf("typed path mask exceeds the source attempt cap")
	}
	document, err := DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	if err != nil {
		return result, nil, err
	}
	prepared, err := document.Prepare()
	if err != nil || int(mask) >= 1<<len(document.Plan.Decisions) {
		return result, nil, fmt.Errorf("typed path candidate exceeds its source palette: %v", err)
	}
	choices := make(map[string]string, len(document.Plan.Decisions))
	for i, choice := range document.Plan.Decisions {
		choices[choice.ID] = choice.Options[(mask>>i)&1].Label
	}
	raw, _ := json.Marshal(document)
	result = PathCandidate{Schema: "gooo/path-candidate/v1", Activity: activity, ActivityID: document.Plan.Base.ID,
		InputSourceSHA256: digest(source), DocumentSHA256: digest(raw), PlanSHA256: prepared.PlanSHA256(),
		Mask: mask, Choices: choices, CandidateSet: set, Stage: "TYPE_CHECK"}
	selected, err := prepared.Compile(choices)
	if err != nil {
		return rejectPathCandidate(ctx, result, err)
	}
	bound, err := bindTypedPathSource(ctx, filename, source, activity, prepared, &BodyPathReceipt{})
	if err != nil {
		return result, nil, err
	}
	result.ActivityID = bound.base.Report.ActivityID
	completed, err := selectedTypedPathSource(source, bound.activity, selected.GoooBody(), choices, true)
	if err != nil {
		return result, nil, err
	}
	fixed, err := fixedCalledSource(filename, string(completed), activity)
	if err != nil {
		return result, nil, err
	}
	projection, err := GenerateWithPlanner(ctx, filename, []byte(fixed), activity, "", "")
	if err != nil || projection.Report.ActivityID != result.ActivityID {
		return result, nil, fmt.Errorf("typed path candidate projection or stable identity differs: %v", err)
	}
	result, err = scoreSourcePathCandidate(ctx, []byte(projection.Source), activity, selected, document.TestCases, result)
	if err != nil {
		return result, nil, err
	}
	result.SelectedSourceSHA256 = digest([]byte(fixed))
	return result, []byte(fixed), nil
}

func scoreSourcePathCandidate(ctx context.Context, projection []byte, activity string,
	selected *bodyplan.Program, tests []pathplan.TestCase, result PathCandidate) (PathCandidate, error) {
	local := make([]IRBodyFillTestCase, len(tests))
	for i, test := range tests {
		local[i] = IRBodyFillTestCase{Input: test.Input, Expected: test.Expected}
	}
	result.Stage = "LOCAL_CASES"
	cases, passed, err := evaluateIntegerCasesContext(ctx, projection, activity, local)
	if err != nil {
		result, _, err = rejectPathCandidate(ctx, result, err)
		return result, err
	}
	for i, test := range tests {
		value, err := selected.Evaluate(test.Input)
		if err != nil || value.Int != cases[i].Actual {
			return result, fmt.Errorf("typed path candidate and emitted evaluator disagree: %v", err)
		}
	}
	result.Stage, result.Cases = "COMPLETE", cases
	result.Passed, result.Total = passed, len(cases)
	return result, nil
}

func rejectPathCandidate(ctx context.Context, result PathCandidate, err error) (PathCandidate, []byte, error) {
	if ctx.Err() != nil {
		return result, nil, ctx.Err()
	}
	result.Failure = err.Error()
	return result, nil, &SourcePathCandidateRejection{cause: err}
}
