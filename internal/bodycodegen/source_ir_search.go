package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// IsSourceIRSearch reports whether an activity owns a generated-candidate IR
// search contract in its Gooo assembling block.
func IsSourceIRSearch(spec *assemblyspec.Spec) bool {
	return spec != nil && spec.Search != nil
}

// GenerateWithSourceIRSearch adapts the source-owned finite contract to the
// same bounded search engine used by external plans.
func GenerateWithSourceIRSearch(ctx context.Context, filename string, source []byte, activity string,
	spec *assemblyspec.Spec, endpoint, apiKey string,
) (Result, error) {
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		return Result{}, err
	}
	return GenerateWithIRBodySearch(ctx, filename, source, activity, plan, endpoint, apiKey)
}

func sourceIRSearchPlan(spec *assemblyspec.Spec) (IRBodySearchPlan, error) {
	if !IsSourceIRSearch(spec) {
		return IRBodySearchPlan{}, fmt.Errorf("activity has no source-declared IR search")
	}
	if err := spec.Validate(); err != nil {
		return IRBodySearchPlan{}, fmt.Errorf("invalid source-declared IR search: %w", err)
	}
	plan := IRBodySearchPlan{
		Schema: bodySearchPlanSchema, Intent: spec.Search.Intent, HoleID: spec.Search.HoleID,
		CandidateGeneration: &IRBodySearchCandidateGeneration{
			Schema: bodySearchCandidateGenerationSchema, Grammar: spec.Search.Grammar,
			MaxCandidates: spec.Search.MaxCandidates,
		},
		TestCases:        make([]IRBodyFillTestCase, len(spec.Cases)),
		HoldoutTestCases: make([]IRBodyFillTestCase, len(spec.HoldoutCases)),
		MaxAttempts:      spec.MaxAttempts,
	}
	for index, testCase := range spec.Cases {
		plan.TestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
	}
	for index, testCase := range spec.HoldoutCases {
		plan.HoldoutTestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
	}
	return plan, nil
}
