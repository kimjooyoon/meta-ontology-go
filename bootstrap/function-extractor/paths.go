package main

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func extractionPath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func extractionLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func editText(lines []string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}

func stageGenericExtraction(root, logical string, buffers map[string][]byte, created map[string]bool, changed, made map[string][]string) error {
	generated, _, err := projectionextractor.Extract(root, logical)
	if err != nil {
		return err
	}
	for path, data := range generated {
		buffers[path] = data
		changed[logical] = appendUnique(changed[logical], path)
		if path != logical {
			created[path] = true
			made[logical] = appendUnique(made[logical], path)
		}
	}
	return nil
}

func extractionFailure(logical string, err error) extractionFailureRecord {
	var failure projectionextractor.Failure
	if errors.As(err, &failure) {
		decision := "UNKNOWN"
		if failure.UnknownClass == "KNOWN_CONTRADICTION" { decision = "REFUTED" }
		return extractionFailureRecord{logical, decision, failure.Stage, failure.Step, failure.Reason, failure.UnknownClass, failure.NextOperation, failure.BlockedBy}
	}
	return extractionFailureRecord{logical, "UNKNOWN", "apply-extraction", "generic", "EXTRACTION_FAILED", "DIRECT_MISSING", "restore-parser-evidence", []string{}}
}
