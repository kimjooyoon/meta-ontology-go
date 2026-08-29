package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func extractionSubjects(plans map[string]planSubject, residual []string, recipes map[string]extractionRecipe, operations map[string][]string, changed, created map[string][]string, staged map[string]stagedFile) ([]extractionSubject, error) {
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
		performed, proof := operations[logical], "coherent-system"
		if recipe, exists := recipes[logical]; exists {
			performed, proof = []string{recipe.Operation}, "axiomatic-foundation"
		}
		if len(performed) == 0 {
			return nil, fmt.Errorf("transformation operation is unavailable for %s", logical)
		}
		result = append(result, extractionSubject{Logical: logical, Before: plans[logical].Lines, After: extractionLines(source.data), Files: files, CreatedFiles: createdFiles, Consumer: "function-extractor", Operation: performed[0], Operations: append([]string{}, performed...), Proof: proof})
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
	temp := name + ".extract.tmp"
	if !sameDirectory(name, temp) {
		return fmt.Errorf("report paths are not same-directory: %s", name)
	}
	if _, err := os.Lstat(temp); !os.IsNotExist(err) {
		return fmt.Errorf("report temporary path exists: %s", temp)
	}
	if err := os.WriteFile(temp, append(encoded, '\n'), 0o644); err != nil {
		return err
	}
	if err := os.Rename(temp, name); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}

func requireHandled(report extractionReport) error {
	for _, failure := range report.Failures {
		if failure.Decision == "REFUTED" {
			return fmt.Errorf("refuted extraction %s: %s", failure.Logical, failure.Reason)
		}
	}
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
