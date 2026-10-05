package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const sourceFillCandidateGenerationReceiptSchema = "gooo/source-fill-candidate-generation-receipt/v1"

// IRBodyFillCandidateGenerationReceipt makes both grammar truncation and
// complete-assignment truncation visible. It measures enumerated space, not intent.
type IRBodyFillCandidateGenerationReceipt struct {
	Schema                         string  `json:"schema"`
	Grammar                        string  `json:"grammar"`
	ExpressionCandidatesEnumerated int     `json:"expression_candidates_enumerated_per_hole"`
	ExpressionsRetainedPerHole     int     `json:"expressions_retained_per_hole"`
	GrammarCoveragePercent         float64 `json:"grammar_coverage_percent"`
	GrammarComplete                bool    `json:"grammar_complete"`
	AssignmentSpaceSize            uint64  `json:"assignment_space_size"`
	AssignmentsRetained            int     `json:"assignments_retained"`
	AssignmentsOmitted             uint64  `json:"assignments_omitted"`
	AssignmentCoveragePercent      float64 `json:"assignment_coverage_percent"`
	CandidateSetSHA256             string  `json:"candidate_set_sha256"`
	CompletenessScope              string  `json:"completeness_scope"`
}

func generateSourceFillCandidates(plan *assemblyspec.FillPlan, cases []assemblyspec.Case) ([]IRBodyFillCandidate, *IRBodyFillCandidateGenerationReceipt, error) {
	if plan == nil || plan.Generation == nil {
		return nil, nil, fmt.Errorf("source fill candidate derivation is not declared")
	}
	search := IRBodySearchPlan{
		TestCases: make([]IRBodyFillTestCase, len(cases)),
		CandidateGeneration: &IRBodySearchCandidateGeneration{
			Schema: bodySearchCandidateGenerationSchema, Grammar: plan.Generation.Grammar,
			MaxCandidates: plan.Generation.MaxExpressions,
		},
	}
	for index, testCase := range cases {
		search.TestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
	}
	grammarReceipt, err := generateIRBodySearchCandidates(&search)
	if err != nil {
		return nil, nil, fmt.Errorf("derive per-hole expression candidates: %w", err)
	}
	expressionCount := len(search.Candidates)
	if expressionCount < 2 {
		return nil, nil, fmt.Errorf("source fill derivation needs at least two expressions per hole")
	}
	spaceSize := uint64(1)
	for range plan.Holes {
		if spaceSize > math.MaxUint64/uint64(expressionCount) {
			return nil, nil, fmt.Errorf("source fill assignment space exceeds the supported counter range")
		}
		spaceSize *= uint64(expressionCount)
	}
	retainedCount := plan.Generation.MaxCandidates
	if spaceSize < uint64(retainedCount) {
		retainedCount = int(spaceSize)
	}
	assignments := make([]IRBodyFillCandidate, 0, retainedCount)
	indices := make([]int, len(plan.Holes))
	for range retainedCount {
		fills := make(map[string]string, len(plan.Holes))
		orderedFills := make([]IRBodyFillHoleFill, len(plan.Holes))
		for holeIndex, hole := range plan.Holes {
			expression := search.Candidates[indices[holeIndex]].Expression
			fills[hole.ID] = expression
			orderedFills[holeIndex] = IRBodyFillHoleFill{HoleID: hole.ID, Expression: expression}
		}
		canonical, _ := json.Marshal(orderedFills)
		digest := sha256.Sum256(canonical)
		assignments = append(assignments, IRBodyFillCandidate{
			ID: "derived_" + hex.EncodeToString(digest[:6]), Fills: fills,
		})
		for index := len(indices) - 1; index >= 0; index-- {
			indices[index]++
			if indices[index] < expressionCount {
				break
			}
			indices[index] = 0
		}
	}
	encoded, err := json.Marshal(assignments)
	if err != nil {
		return nil, nil, fmt.Errorf("encode derived source fill assignments: %w", err)
	}
	digest := sha256.Sum256(encoded)
	coverage := float64(len(assignments)) * 100 / float64(spaceSize)
	receipt := &IRBodyFillCandidateGenerationReceipt{
		Schema: sourceFillCandidateGenerationReceiptSchema, Grammar: plan.Generation.Grammar,
		ExpressionCandidatesEnumerated: grammarReceipt.CandidatesEnumerated,
		ExpressionsRetainedPerHole:     expressionCount, GrammarCoveragePercent: grammarReceipt.GrammarCoveragePercent,
		GrammarComplete: grammarReceipt.GrammarComplete, AssignmentSpaceSize: spaceSize,
		AssignmentsRetained: len(assignments), AssignmentsOmitted: spaceSize - uint64(len(assignments)),
		AssignmentCoveragePercent: coverage, CandidateSetSHA256: "sha256:" + hex.EncodeToString(digest[:]),
		CompletenessScope: "declared per-hole integer-offset-constant/v1 expressions and lexicographically enumerated complete assignments under the source candidate cap; excludes other grammars, later assignments, unstated intent and all-domain semantics",
	}
	return assignments, receipt, nil
}
