package languagediagnosticprovenancebinding

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
)

func coordinates(input Input) []Coordinate {
	classes := indicatorClasses(input.Provenance)
	return []Coordinate{
		coordinate("exact-head", "non-empty", emptyLabel(input.ExpectedHeadSHA),
			input.ExpectedHeadSHA != "" && input.Provenance.Source.ExpectedHeadSHA == input.ExpectedHeadSHA),
		coordinate("concept-artifact", "PASS", input.Concept.Decision,
			input.Concept.Decision == "PASS" && input.Concept.Ready()),
		coordinate("concept-count", ">=18", fmt.Sprint(input.Concept.Report.Summary.Concepts),
			input.Concept.Report.Summary.Concepts >= ExpectedConcepts),
		coordinate("provenance-decision", "PASS/EXACT",
			string(input.Provenance.Decision)+"/"+string(input.Provenance.Resolution),
			input.Provenance.Decision == languagediagnosticprovenance.DecisionPass &&
				input.Provenance.Resolution == languagediagnosticprovenance.ResolutionExact),
		coordinate("provenance-corpus", "18/18",
			fmt.Sprintf("%d/%d", input.Provenance.Summary.Satisfied, input.Provenance.Summary.Total),
			input.Provenance.Summary.Satisfied == 18 && input.Provenance.Summary.Total == 18),
		coordinate("provenance-indicators", "18:3/8/7",
			fmt.Sprintf("%d:%d/%d/%d", len(input.Provenance.Indicators), classes[0], classes[1], classes[2]),
			len(input.Provenance.Indicators) == 18 && classes == [3]int{3, 8, 7}),
		coordinate("readiness-contract", "gooo/self-improving-language-obligations/v1:24",
			input.Readiness.ContractSchema+":"+fmt.Sprint(input.Readiness.Summary.Total),
			input.Readiness.ContractSchema == "gooo/self-improving-language-obligations/v1" &&
				input.Readiness.Summary.Total == 24),
		coordinate("provenance-obligation", "SATISFIED", obligationStatus(input),
			obligationStatus(input) == "SATISFIED"),
		coordinate("readiness-count", ">=18/24",
			fmt.Sprintf("%d/%d", input.Readiness.Summary.Completed, input.Readiness.Summary.Total),
			input.Readiness.Summary.Completed >= 18 && input.Readiness.Summary.Total == 24),
		coordinate("readiness-bps", ">=7500", fmt.Sprint(input.Readiness.Summary.ReadinessBPS),
			input.Readiness.Summary.ReadinessBPS >= 7500),
		coordinate("provenance-proofs", "3/3", proofLabel(input), proofsBound(input)),
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
