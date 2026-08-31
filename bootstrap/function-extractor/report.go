package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	markSubjectState(subjects, "STAGED_NOT_COMMITTED")
	return extractionReport{
		Schema: "gooo.function-extraction.v2", SourceSHA: sha, StagedSubjects: len(subjects),
		Subjects: subjects, Unhandled: unhandled, Failures: failures,
		Indicators: []extractionIndicator{
			{ID: "extraction.observed", Value: observed, Limit: -1,
				Consumer: "function-extractor", Operation: "observe-density-residual", Proof: "axiomatic-foundation"},
			{ID: "extraction.staged", Value: len(subjects), Limit: -1,
				Consumer: "function-extractor", Operation: "stage-helper-extraction", Proof: "coherent-system"},
			{ID: "extraction.applied", Value: 0, Limit: -1,
				Consumer: "logical-materializer", Operation: "accept-helper-extraction", Proof: "coherent-system"},
			{ID: "extraction.created", Value: 0, Limit: -1,
				Consumer: "authorized-write-set", Operation: "authorize-declared-file-creation", Proof: "axiomatic-foundation"},
			{ID: "extraction.unhandled", Value: len(unhandled), Limit: 0, Blocking: true,
				Consumer: "function-extractor", Operation: "define-extraction-recipe", Proof: "infinite-regress"},
		},
	}
}

func markSubjectState(subjects []extractionSubject, state string) {
	for index := range subjects {
		subjects[index].State = state
	}
}

func markCommitted(report *extractionReport) {
	markSubjectState(report.Subjects, "COMMITTED_APPLIED")
	for index := range report.Indicators {
		switch report.Indicators[index].ID {
		case "extraction.applied":
			report.Indicators[index].Value = len(report.Subjects)
		case "extraction.created":
			report.Indicators[index].Value = createdCount(report.Subjects)
		}
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
	file, err := os.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		_ = file.Close()
		_ = os.Remove(temp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(temp)
		return err
	}
	if err := os.Rename(temp, name); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}

func createProvisionalReportPath(name string, report extractionReport) (string, error) {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp(filepath.Dir(name), ".extract-report-provisional-*")
	if err != nil {
		return "", err
	}
	path := file.Name()
	payload := append(encoded, '\n')
	written, err := file.Write(payload)
	if err != nil || written != len(payload) {
		_ = file.Close()
		_ = os.Remove(path)
		if err == nil {
			err = fmt.Errorf("short provisional report write: %d/%d", written, len(payload))
		}
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func removeReport(name string) error {
	if name == "" {
		return nil
	}
	err := os.Remove(name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
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
