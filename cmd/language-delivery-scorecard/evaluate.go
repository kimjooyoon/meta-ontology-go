package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagedelivery"
)

func evaluate(value options, stdout, stderr io.Writer) int {
	contract, err := languagedelivery.ReadContract(value.contract)
	if err != nil {
		fmt.Fprintf(stderr, "language-delivery-scorecard: contract: %v\n", err)
		return 2
	}
	manifest, err := languagedelivery.ReadManifest(value.manifest)
	if err != nil {
		fmt.Fprintf(stderr, "language-delivery-scorecard: manifest: %v\n", err)
		return 2
	}
	paths := map[languagedelivery.SourceName]string{
		languagedelivery.SourceUserJourney: value.user,
		languagedelivery.SourceConformance: value.conformance,
		languagedelivery.SourceLSP:         value.lsp,
		languagedelivery.SourceRelease:     value.release,
		languagedelivery.SourceExecution:   value.execution,
		languagedelivery.SourceTest:        value.languageTest,
		languagedelivery.SourceProfile:     value.profile,
		languagedelivery.SourceDebug:       value.debug,
		languagedelivery.SourceReadiness:   value.readiness,
	}
	evidence, err := languagedelivery.ReadEvidence(paths)
	if err != nil {
		fmt.Fprintf(stderr, "language-delivery-scorecard: evidence: %v\n", err)
		return 2
	}
	report, err := languagedelivery.Evaluate(contract, manifest, evidence, value.head)
	if err != nil {
		fmt.Fprintf(stderr, "language-delivery-scorecard: evaluate: %v\n", err)
		return 2
	}
	if err := languagedelivery.WriteReport(value.out, report); err != nil {
		fmt.Fprintf(stderr, "language-delivery-scorecard: write: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "language delivery: %s %d/%d (%d bps)\n", report.Decision,
		report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total, report.Summary.Coordinates.BasisPoints)
	if report.Decision == "FAIL_CLOSED" {
		return 1
	}
	return 0
}
