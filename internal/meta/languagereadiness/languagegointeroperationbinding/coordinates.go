package languagegointeroperationbinding

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagegointeroperation"
)

func coordinates(input Input) []Coordinate {
	classes := indicatorClasses(input.Interoperation)
	return []Coordinate{
		coordinate("exact-head", "non-empty", emptyLabel(input.ExpectedHeadSHA), input.ExpectedHeadSHA != "" && input.Interoperation.Source.ExpectedHeadSHA == input.ExpectedHeadSHA),
		coordinate("concept-artifact", "PASS", input.Concept.Decision, input.Concept.Decision == "PASS" && input.Concept.Ready()),
		coordinate("concept-count", "17", fmt.Sprint(input.Concept.Report.Summary.Concepts), input.Concept.Report.Summary.Concepts == ExpectedConcepts),
		coordinate("interop-decision", "PASS/EXACT", string(input.Interoperation.Decision)+"/"+string(input.Interoperation.Resolution), input.Interoperation.Decision == languagegointeroperation.DecisionPass && input.Interoperation.Resolution == languagegointeroperation.ResolutionExact),
		coordinate("interop-corpus", "24/24", fmt.Sprintf("%d/%d", input.Interoperation.Summary.Satisfied, input.Interoperation.Summary.Total), input.Interoperation.Summary.Satisfied == 24 && input.Interoperation.Summary.Total == 24),
		coordinate("interop-indicators", "18:3/8/7", fmt.Sprintf("%d:%d/%d/%d", len(input.Interoperation.Indicators), classes[0], classes[1], classes[2]), len(input.Interoperation.Indicators) == 18 && classes == [3]int{3, 8, 7}),
		coordinate("readiness-contract", "gooo/self-improving-language-obligations/v1:24", input.Readiness.ContractSchema+":"+fmt.Sprint(input.Readiness.Summary.Total), input.Readiness.ContractSchema == "gooo/self-improving-language-obligations/v1" && input.Readiness.Summary.Total == 24),
		coordinate("interop-obligation", "SATISFIED", interopObligationStatus(input), interopObligationStatus(input) == "SATISFIED"),
		coordinate("readiness-count", "17/24", fmt.Sprintf("%d/%d", input.Readiness.Summary.Completed, input.Readiness.Summary.Total), input.Readiness.Summary.Completed == 17 && input.Readiness.Summary.Total == 24),
		coordinate("readiness-bps", "7083", fmt.Sprint(input.Readiness.Summary.ReadinessBPS), input.Readiness.Summary.ReadinessBPS == 7083),
		coordinate("interop-proofs", "3/3", proofLabel(input), proofsBound(input)),
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
