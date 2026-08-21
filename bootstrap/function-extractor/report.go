package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func extractionEvidence(sha string, subjects []extractionSubject,
	unhandled []string) extractionReport {
	proof := "axiomatic-foundation"
	return extractionReport{
		Schema: "gooo.function-extraction.v1", SourceSHA: sha,
		Subjects: subjects, Unhandled: unhandled,
		Indicators: []extractionIndicator{
			{ID: "extraction.applied", Value: len(subjects), Limit: -1,
				Consumer: "logical-materializer", Operation: "accept-helper-extraction", Proof: proof},
			{ID: "extraction.unhandled", Value: len(unhandled), Limit: 0, Blocking: true,
				Consumer: "function-extractor", Operation: "define-extraction-recipe", Proof: proof},
		},
	}
}

func writeExtractionReport(name string, report extractionReport) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(encoded, '\n'), 0o644)
}

func requireHandled(report extractionReport) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
