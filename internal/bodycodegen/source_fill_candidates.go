package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const sourceFillCandidateGenerationReceiptSchema = "gooo/source-fill-candidate-generation-receipt/v2"

type IRBodyFillHoleGrammarCoverage struct {
	HoleID                    string  `json:"hole_id"`
	Grammar                   string  `json:"grammar"`
	ExpressionCandidatesTotal int     `json:"expression_candidates_total"`
	ExpressionsRetained       int     `json:"expressions_retained"`
	GrammarCoveragePercent    float64 `json:"grammar_coverage_percent"`
	GrammarComplete           bool    `json:"grammar_complete"`
}

// IRBodyFillCandidateGenerationReceipt measures the per-hole grammar spaces and
// the complete assignments that were enumerated from them. It does not measure intent.
type IRBodyFillCandidateGenerationReceipt struct {
	Schema                    string                          `json:"schema"`
	Grammar                   string                          `json:"grammar"`
	HoleGrammars              []IRBodyFillHoleGrammarCoverage `json:"hole_grammars"`
	ExpressionCandidatesTotal int                             `json:"expression_candidates_total"`
	ExpressionsRetainedTotal  int                             `json:"expressions_retained_total"`
	GrammarCoveragePercent    float64                         `json:"grammar_coverage_percent"`
	GrammarComplete           bool                            `json:"grammar_complete"`
	AssignmentSpaceSize       uint64                          `json:"assignment_space_size"`
	AssignmentsRetained       int                             `json:"assignments_retained"`
	AssignmentsOmitted        uint64                          `json:"assignments_omitted"`
	AssignmentCoveragePercent float64                         `json:"assignment_coverage_percent"`
	CandidateSetSHA256        string                          `json:"candidate_set_sha256"`
	CompletenessScope         string                          `json:"completeness_scope"`
}

func generateSourceFillCandidates(plan *assemblyspec.FillPlan, cases []assemblyspec.Case) ([]IRBodyFillCandidate, *IRBodyFillCandidateGenerationReceipt, error) {
	if plan == nil || plan.Generation == nil {
		return nil, nil, fmt.Errorf("source fill candidate derivation is not declared")
	}
	grammarByHole := make(map[string]assemblyspec.FillHoleGrammar, len(plan.Holes))
	if len(plan.Generation.HoleGrammars) != 0 {
		for _, grammar := range plan.Generation.HoleGrammars {
			grammarByHole[grammar.HoleID] = grammar
		}
	} else {
		for _, hole := range plan.Holes {
			grammarByHole[hole.ID] = assemblyspec.FillHoleGrammar{
				HoleID: hole.ID, Grammar: plan.Generation.Grammar,
				MaxExpressions: plan.Generation.MaxExpressions,
			}
		}
	}
	expressions := make([][]string, len(plan.Holes))
	holeCoverage := make([]IRBodyFillHoleGrammarCoverage, len(plan.Holes))
	var expressionTotal, retainedTotal int
	grammarComplete := true
	commonGrammar := ""
	for holeIndex, hole := range plan.Holes {
		declaration, ok := grammarByHole[hole.ID]
		if !ok {
			return nil, nil, fmt.Errorf("source fill has no grammar for hole %q", hole.ID)
		}
		candidates, enumerated, complete, err := generateSourceFillExpressions(declaration, cases)
		if err != nil {
			return nil, nil, fmt.Errorf("derive expressions for hole %q: %w", hole.ID, err)
		}
		if len(candidates) < 2 {
			return nil, nil, fmt.Errorf("source fill derivation needs at least two expressions for hole %q", hole.ID)
		}
		expressions[holeIndex] = candidates
		coverage := float64(len(candidates)) * 100 / float64(enumerated)
		holeCoverage[holeIndex] = IRBodyFillHoleGrammarCoverage{
			HoleID: hole.ID, Grammar: declaration.Grammar, ExpressionCandidatesTotal: enumerated,
			ExpressionsRetained: len(candidates), GrammarCoveragePercent: coverage,
			GrammarComplete: complete,
		}
		expressionTotal += enumerated
		retainedTotal += len(candidates)
		grammarComplete = grammarComplete && complete
		if holeIndex == 0 {
			commonGrammar = declaration.Grammar
		} else if commonGrammar != declaration.Grammar {
			commonGrammar = "per-hole"
		}
	}
	spaceSize := uint64(1)
	for _, candidates := range expressions {
		if spaceSize > math.MaxUint64/uint64(len(candidates)) {
			return nil, nil, fmt.Errorf("source fill assignment space exceeds the supported counter range")
		}
		spaceSize *= uint64(len(candidates))
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
			expression := expressions[holeIndex][indices[holeIndex]]
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
			if indices[index] < len(expressions[index]) {
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
	grammarCoverage := float64(retainedTotal) * 100 / float64(expressionTotal)
	receipt := &IRBodyFillCandidateGenerationReceipt{
		Schema: sourceFillCandidateGenerationReceiptSchema, Grammar: commonGrammar,
		HoleGrammars: holeCoverage, ExpressionCandidatesTotal: expressionTotal,
		ExpressionsRetainedTotal: retainedTotal, GrammarCoveragePercent: grammarCoverage,
		GrammarComplete: grammarComplete, AssignmentSpaceSize: spaceSize,
		AssignmentsRetained: len(assignments), AssignmentsOmitted: spaceSize - uint64(len(assignments)),
		AssignmentCoveragePercent: coverage, CandidateSetSHA256: "sha256:" + hex.EncodeToString(digest[:]),
		CompletenessScope: "declared per-hole typed expression grammars and lexicographically enumerated complete assignments under the source candidate cap; excludes other grammars, later assignments, unstated intent and all-domain semantics",
	}
	return assignments, receipt, nil
}

func generateSourceFillExpressions(declaration assemblyspec.FillHoleGrammar, cases []assemblyspec.Case) ([]string, int, bool, error) {
	switch declaration.Grammar {
	case "integer-offset-constant/v1":
		search := IRBodySearchPlan{
			TestCases: make([]IRBodyFillTestCase, len(cases)),
			CandidateGeneration: &IRBodySearchCandidateGeneration{
				Schema: bodySearchCandidateGenerationSchema, Grammar: declaration.Grammar,
				MaxCandidates: declaration.MaxExpressions,
			},
		}
		for index, testCase := range cases {
			search.TestCases[index] = IRBodyFillTestCase{Input: testCase.Input, Expected: testCase.Expected}
		}
		receipt, err := generateIRBodySearchCandidates(&search)
		if err != nil {
			return nil, 0, false, err
		}
		expressions := make([]string, len(search.Candidates))
		for index, candidate := range search.Candidates {
			expressions[index] = candidate.Expression
		}
		return expressions, receipt.CandidatesEnumerated, receipt.GrammarComplete, nil
	case "integer-predicate/v1":
		expressions := []string{"true", "false"}
		seen := map[string]bool{"true": true, "false": true}
		inputs := make([]int64, len(cases))
		for index, testCase := range cases {
			inputs[index] = testCase.Input
		}
		slices.Sort(inputs)
		uniqueInputs := slices.Compact(inputs)
		for _, operator := range []string{"<", "<=", ">", ">=", "==", "!="} {
			for _, input := range uniqueInputs {
				literal := strconv.FormatInt(input, 10)
				expression := "input " + operator + " " + literal
				if !seen[expression] {
					seen[expression] = true
					expressions = append(expressions, expression)
				}
			}
		}
		enumerated := len(expressions)
		if len(expressions) > declaration.MaxExpressions {
			expressions = expressions[:declaration.MaxExpressions]
		}
		return expressions, enumerated, len(expressions) == enumerated, nil
	default:
		return nil, 0, false, fmt.Errorf("unsupported source-fill grammar %q", declaration.Grammar)
	}
}
