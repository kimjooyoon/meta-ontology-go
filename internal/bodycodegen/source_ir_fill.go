package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// IsSourceIRBodyFill reports whether the activity owns a complete multi-hole
// body-fill plan in its Gooo assembling block.
func IsSourceIRBodyFill(spec *assemblyspec.Spec) bool {
	return spec != nil && spec.FillPlan != nil
}

// GenerateWithSourceIRBodyFill projects source-owned hole assignments through
// the same typed scoring and synchronous provider path as an external plan.
func GenerateWithSourceIRBodyFill(ctx context.Context, filename string, source []byte, activity string,
	spec *assemblyspec.Spec, endpoint, apiKey string, options IRBodyFillOptions,
) (Result, error) {
	if !IsSourceIRBodyFill(spec) {
		return Result{}, fmt.Errorf("activity has no source-declared IR body-fill plan")
	}
	if err := spec.Validate(); err != nil {
		return Result{}, fmt.Errorf("invalid source-declared IR body-fill plan: %w", err)
	}
	candidates := make([]IRBodyFillCandidate, len(spec.FillPlan.Candidates))
	var generationReceipt *IRBodyFillCandidateGenerationReceipt
	if spec.FillPlan.Generation != nil {
		var err error
		candidates, generationReceipt, err = generateSourceFillCandidates(spec.FillPlan, spec.Cases)
		if err != nil {
			return Result{}, err
		}
	}
	if spec.FillPlan.Generation == nil {
		for index, candidate := range spec.FillPlan.Candidates {
			fills := make(map[string]string, len(candidate.Fills))
			for _, fill := range candidate.Fills {
				fills[fill.HoleID] = fill.Expression
			}
			candidates[index] = IRBodyFillCandidate{ID: candidate.ID, Fills: fills}
		}
	}
	plan := IRBodyFillPlan{
		Schema: bodyFillMultiPlanSchema, Intent: spec.FillPlan.Intent,
		Holes:            make([]IRBodyFillHole, len(spec.FillPlan.Holes)),
		Candidates:       make([]IRBodyFillCandidate, len(candidates)),
		TestCases:        make([]IRBodyFillTestCase, len(spec.Cases)),
		HoldoutTestCases: make([]IRBodyFillTestCase, len(spec.HoldoutCases)),
	}
	for index, hole := range spec.FillPlan.Holes {
		plan.Holes[index] = IRBodyFillHole{ID: hole.ID}
	}
	for index, candidate := range candidates {
		plan.Candidates[index] = candidate
	}
	for index, testCase := range spec.Cases {
		plan.TestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
	}
	for index, testCase := range spec.HoldoutCases {
		plan.HoldoutTestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
	}
	result, err := GenerateWithIRBodyFillWithOptions(ctx, filename, source, activity, plan, endpoint, apiKey, options)
	if err != nil {
		return Result{}, err
	}
	var output []byte
	output, err = sourceWithoutIRBodyFill(filename, []byte(result.GoooSource), activity)
	if err != nil {
		return Result{}, err
	}
	result.GoooSource = string(output)
	if generationReceipt != nil && result.Report.BodyFill != nil {
		result.Report.BodyFill.CandidateGeneration = generationReceipt
		populateCompletenessReceipt(&result.Report, "")
	}
	return result, nil
}

func sourceWithoutIRBodyFill(filename string, source []byte, activity string) ([]byte, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("source body-fill output: %v", diagnostics)
	}
	for _, declaration := range file.Declarations {
		selected, ok := declaration.(*syntax.ActivityDecl)
		if !ok || selected.Name != activity || selected.Assembly == nil || selected.Assembly.Spec.FillPlan == nil {
			continue
		}
		start, end := selected.Assembly.Span.Start.Offset, selected.Assembly.Span.End.Offset
		if start < selected.ValueProgramSpan.End.Offset || end < start || end > len(source) {
			return nil, fmt.Errorf("source body-fill spans are inconsistent")
		}
		if start > 0 && source[start-1] == ' ' {
			start--
		}
		updated := make([]byte, 0, len(source)-(end-start))
		updated = append(updated, source[:start]...)
		updated = append(updated, source[end:]...)
		return updated, nil
	}
	return nil, fmt.Errorf("source body-fill activity %q was not found", activity)
}
