package languagedeterministicquerybinding

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagedeterministicquery"
)

func coordinates(input Input) []Coordinate {
	queryIndicators := indicatorClasses(input.Query)
	return []Coordinate{
		coordinate("exact-head", "non-empty", emptyLabel(input.ExpectedHeadSHA), input.ExpectedHeadSHA != "" && input.Query.Source.ExpectedHeadSHA == input.ExpectedHeadSHA),
		coordinate("concept-artifact", "PASS", input.Concept.Decision, input.Concept.Decision == "PASS" && input.Concept.Ready()),
		coordinate("concept-count", ">=16", fmt.Sprint(input.Concept.Report.Summary.Concepts), input.Concept.Report.Summary.Concepts >= ExpectedConcepts),
		coordinate("query-decision", "PASS/EXACT", string(input.Query.Decision)+"/"+string(input.Query.Resolution), input.Query.Decision == languagedeterministicquery.DecisionPass && input.Query.Resolution == languagedeterministicquery.ResolutionExact),
		coordinate("query-corpus", "32/32", fmt.Sprintf("%d/%d", input.Query.Summary.Satisfied, input.Query.Summary.Total), input.Query.Summary.Satisfied == ExpectedQueryCases && input.Query.Summary.Total == ExpectedQueryCases),
		coordinate("query-indicators", "18:1/8/9", fmt.Sprintf("%d:%d/%d/%d", len(input.Query.Indicators), queryIndicators[0], queryIndicators[1], queryIndicators[2]), len(input.Query.Indicators) == ExpectedIndicators && queryIndicators == [3]int{1, 8, 9}),
		coordinate("readiness-contract", "gooo/self-improving-language-obligations/v1:24", input.Readiness.ContractSchema+":"+fmt.Sprint(input.Readiness.Summary.Total), input.Readiness.ContractSchema == "gooo/self-improving-language-obligations/v1" && input.Readiness.Summary.Total == ExpectedTotal),
		coordinate("query-obligation", "SATISFIED", queryObligationStatus(input), queryObligationStatus(input) == "SATISFIED"),
		coordinate("readiness-count", ">=16/24", fmt.Sprintf("%d/%d", input.Readiness.Summary.Completed, input.Readiness.Summary.Total), input.Readiness.Summary.Completed >= ExpectedCompleted && input.Readiness.Summary.Total == ExpectedTotal),
		coordinate("readiness-bps", ">=6666", fmt.Sprint(input.Readiness.Summary.ReadinessBPS), input.Readiness.Summary.ReadinessBPS >= ExpectedBPS),
		coordinate("query-proofs", "3/3", queryProofLabel(input), queryProofsBound(input)),
		coordinate("sealed-effects", "0/0/0/0", sealedLabel(input), sealedBound(input)),
	}
}

func coordinate(id, expected, observed string, bound bool) Coordinate {
	return Coordinate{ID: id, Expected: expected, Observed: observed, Bound: bound}
}

func emptyLabel(value string) string {
	if value == "" {
		return "empty"
	}
	return "non-empty"
}
