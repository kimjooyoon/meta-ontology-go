package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func extractionSubjects(plans map[string]planSubject, residual []string, recipes map[string]extractionRecipe, changed, created map[string][]string, staged map[string]stagedFile) ([]extractionSubject, error) {
	result := make([]extractionSubject, 0, len(residual))
	for _, logical := range residual {
		files := changed[logical]
		if len(files) == 0 {
			continue
		}
		sort.Strings(files)
		createdFiles := created[logical]
		sort.Strings(createdFiles)
		source, exists := staged[logical]
		if !exists {
			return nil, fmt.Errorf("recipe did not rewrite subject %s", logical)
		}
		operation, proof := "extract-function", "coherent-system"
		if recipe, exists := recipes[logical]; exists {
			operation, proof = recipe.Operation, "axiomatic-foundation"
		}
		result = append(result, extractionSubject{Logical: logical, Before: plans[logical].Lines, After: extractionLines(source.data), Files: files, CreatedFiles: createdFiles, Consumer: "function-extractor", Operation: operation, Proof: proof})
	}
	return result, nil
}

func extractionEvidence(sha string, subjects []extractionSubject,
	unhandled []string, failures []extractionFailureRecord) extractionReport {
	observed := len(subjects) + len(unhandled)
	created := createdCount(subjects)
	return extractionReport{
		Schema: "gooo.function-extraction.v1", SourceSHA: sha,
		Subjects: subjects, Unhandled: unhandled, Failures: failures,
		Indicators: []extractionIndicator{
			{ID: "extraction.observed", Value: observed, Limit: -1,
				Consumer: "function-extractor", Operation: "observe-density-residual", Proof: "axiomatic-foundation"},
			{ID: "extraction.applied", Value: len(subjects), Limit: -1,
				Consumer: "logical-materializer", Operation: "accept-helper-extraction", Proof: "coherent-system"},
			{ID: "extraction.created", Value: created, Limit: -1,
				Consumer: "authorized-write-set", Operation: "authorize-declared-file-creation", Proof: "axiomatic-foundation"},
			{ID: "extraction.unhandled", Value: len(unhandled), Limit: 0, Blocking: true,
				Consumer: "function-extractor", Operation: "define-extraction-recipe", Proof: "infinite-regress"},
		},
	}
}

func createdCount(subjects []extractionSubject) int {
	count := 0
	for _, subject := range subjects {
		count += len(subject.CreatedFiles)
	}
	return count
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
