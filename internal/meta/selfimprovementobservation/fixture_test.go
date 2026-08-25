package selfimprovementobservation

import (
	"fmt"
	"strings"
)

func validFixture() (Inputs, Options) {
	head, raw := strings.Repeat("a", 40), strings.Repeat("1", 64)
	facts := "sha256:" + raw
	indicators := make([]SourceIndicator, 15)
	for index := range indicators {
		choice := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}[index%3]
		indicators[index] = SourceIndicator{ID: fmt.Sprintf("source.%02d", index+1), Class: "DRIVER", ProofChoice: choice, MetaOperation: fmt.Sprintf("operation-%02d", index+1), Value: 1, Target: 1, Satisfied: true}
	}
	report := SourceReport{
		Schema: "gooo/language-example-experiment-report/v2", Decision: "PASS", Resolution: "EXACT",
		Reason: "EXPERIMENT_CONTRACT_OBSERVED", Interpretation: "MINIMAL_VALUE_OBSERVED",
		SubjectSHA: head, ContractID: "billing-operation-manifest-v2",
		Summary: SourceSummary{
			Coordinates: CountSummary{15, 15, 10000}, Value: SourceValue{1, 3, 1, 1},
			Compiler: SourceCompiler{2, 2, 0, 10000, 3},
			Resources: SourceResources{5, 5, 5, 10724, 12529067, 0, 0, 0},
			Counterexamples: SourceCounterexamples{1}, Effects: SourceEffects{}, NotClaimed: 5,
		},
		Indicators: indicators, Views: sourceViews(indicators),
		Proofs: []SourceProof{{"FOUNDATION", "foundation", "foundation-op", facts, true}, {"COHERENCE", "coherence", "coherence-op", facts, true}, {"REGRESSION", "regression", "regression-op", facts, true}},
		NotClaimed: []string{"business correctness", "value-level computation", "production readiness", "performance beyond this runner and fixed sample set", "general-purpose code generation"},
		FactsDigest: facts,
	}
	report.Digest = digestJSON(report)
	contract := contractFixture(head, raw)
	docDigest := "sha256:" + strings.Repeat("2", 64)
	return Inputs{Document[SourceReport]{report, docDigest}, Document[CounterexampleReport]{CounterexampleReport{"gooo/language-example-counterexamples/v1", 6, 6}, docDigest}, Document[ContractReport]{contract, docDigest}}, Options{head, 42}
}

func sourceViews(indicators []SourceIndicator) []SourceView {
	ids := make([]string, len(indicators))
	for index := range indicators {
		ids[index] = indicators[index].ID
	}
	return []SourceView{{"USER", "USER_VISIBLE", 6, 6, 10000, ids[:6]}, {"TOOL_AUTHOR", "TOOL_CONTRACT", 12, 12, 10000, ids[:12]}, {"GOVERNOR", "FULL_RECEIPT", 15, 15, 10000, ids}}
}
