package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

const bodySearchCandidateGenerationSchema = "gooo/body-codegen-ir-candidate-generation/v1"
const bodySearchCandidateGenerationReceiptSchema = "gooo/body-codegen-ir-candidate-generation-receipt/v1"
const bodySearchIntegerAffineGrammar = "integer-offset-constant/v1"
const bodySearchHoleResidualGrammar = "integer-hole-residual/v1"

// IRBodySearchCandidateGeneration asks Gooo to derive a bounded expression set
// from the declared training examples instead of requiring a hand-written list.
type IRBodySearchCandidateGeneration struct {
	Schema        string `json:"schema"`
	Grammar       string `json:"grammar"`
	MaxCandidates int    `json:"max_candidates"`
}

// IRBodySearchCandidateGenerationReceipt states the exact boundary of the
// generated candidate space. Its percentage is grammar coverage, not intent
// accuracy or whole-domain completeness.
type IRBodySearchCandidateGenerationReceipt struct {
	Schema                 string             `json:"schema"`
	Grammar                string             `json:"grammar"`
	CandidatesEnumerated   int                `json:"candidates_enumerated"`
	CandidatesRetained     int                `json:"candidates_retained"`
	CandidatesOmitted      int                `json:"candidates_omitted"`
	CandidateSetSHA256     string             `json:"candidate_set_sha256"`
	GrammarCoveragePercent float64            `json:"grammar_coverage_percent"`
	GrammarComplete        bool               `json:"grammar_complete"`
	CompletenessScope      string             `json:"completeness_scope"`
	HoleContext            *IRBodyHoleContext `json:"hole_context,omitempty"`
}

func validateIRBodySearchCandidateGeneration(plan IRBodySearchCandidateGeneration) error {
	if plan.Schema != bodySearchCandidateGenerationSchema {
		return fmt.Errorf("IR body-search candidate-generation schema must be %q", bodySearchCandidateGenerationSchema)
	}
	if plan.Grammar != bodySearchIntegerAffineGrammar && plan.Grammar != bodySearchHoleResidualGrammar {
		return fmt.Errorf("unsupported IR body-search candidate grammar %q", plan.Grammar)
	}
	if plan.MaxCandidates < 2 || plan.MaxCandidates > 16 {
		return fmt.Errorf("candidate_generation.max_candidates must be 2..16")
	}
	return nil
}

func generateIRBodySearchCandidates(plan *IRBodySearchPlan) (*IRBodySearchCandidateGenerationReceipt, error) {
	if plan.CandidateGeneration == nil {
		return nil, nil
	}
	if err := validateIRBodySearchCandidateGeneration(*plan.CandidateGeneration); err != nil {
		return nil, err
	}
	if plan.CandidateGeneration.Grammar != bodySearchIntegerAffineGrammar {
		return nil, fmt.Errorf("hole residual grammar requires source context")
	}
	return retainIRBodySearchCandidates(plan, bodySearchAffineExpressions(plan.TestCases), nil)
}

func bodySearchAffineExpressions(cases []IRBodyFillTestCase) []string {
	expressions := make([]string, 0, 4+len(cases)*3)
	seen := make(map[string]bool, cap(expressions))
	add := func(expression string) {
		if expression != "" && !seen[expression] {
			seen[expression] = true
			expressions = append(expressions, expression)
		}
	}
	add("input")
	for _, testCase := range cases {
		add(strconv.FormatInt(testCase.Expected, 10))
	}
	add("-input")
	for _, testCase := range cases {
		delta, ok := bodySearchInt64Delta(testCase.Input, testCase.Expected)
		if !ok {
			continue
		}
		add("input + " + strconv.FormatInt(delta, 10))
		if delta != -1<<63 {
			add("input - " + strconv.FormatInt(-delta, 10))
		}
	}
	return expressions
}

func retainIRBodySearchCandidates(plan *IRBodySearchPlan, expressions []string, holeContext *IRBodyHoleContext) (*IRBodySearchCandidateGenerationReceipt, error) {
	if len(expressions) < 2 {
		return nil, fmt.Errorf("candidate generation produced fewer than two distinct expressions")
	}
	retained := expressions
	if len(retained) > plan.CandidateGeneration.MaxCandidates {
		retained = retained[:plan.CandidateGeneration.MaxCandidates]
	}
	candidates := make([]IRBodyFillCandidate, 0, len(retained))
	for _, expression := range retained {
		candidates = append(candidates, IRBodyFillCandidate{ID: generatedSearchCandidateID(expression), Expression: expression})
	}
	plan.Candidates = candidates
	encoded, err := json.Marshal(candidates)
	if err != nil {
		return nil, fmt.Errorf("encode generated body-search candidates: %w", err)
	}
	digest := sha256.Sum256(encoded)
	coverage := float64(len(retained)) * 100 / float64(len(expressions))
	omitted := len(expressions) - len(retained)
	scope := "retained unique expressions / all unique expressions in integer-offset-constant/v1 derived from training examples; excludes other expression grammars and all-domain semantics"
	if holeContext != nil {
		scope = "retained unique expressions / contextual constants, input offsets, first-anchor affine fits and legacy grammar seeds derived from training-only zero/one hole probes; excludes other grammars and all-domain semantics"
	}
	return &IRBodySearchCandidateGenerationReceipt{
		Schema: bodySearchCandidateGenerationReceiptSchema, Grammar: plan.CandidateGeneration.Grammar,
		CandidatesEnumerated: len(expressions), CandidatesRetained: len(retained), CandidatesOmitted: omitted,
		CandidateSetSHA256: "sha256:" + hex.EncodeToString(digest[:]), GrammarCoveragePercent: coverage,
		GrammarComplete:   omitted == 0,
		CompletenessScope: scope, HoleContext: holeContext,
	}, nil
}

func generatedSearchCandidateID(expression string) string {
	digest := sha256.Sum256([]byte(expression))
	return "generated_" + hex.EncodeToString(digest[:6])
}

func bodySearchInt64Delta(input, expected int64) (int64, bool) {
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -1 << 63
	if input > 0 && expected < minInt64+input || input < 0 && expected > maxInt64+input {
		return 0, false
	}
	return expected - input, true
}
