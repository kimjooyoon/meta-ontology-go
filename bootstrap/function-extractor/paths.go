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
	if failure, ok := errors.AsType[projectionextractor.Failure](err); ok {
		decision := "UNKNOWN"
		if failure.UnknownClass == "KNOWN_CONTRADICTION" {
			decision = "REFUTED"
		}
		return extractionFailureRecord{Logical: logical, BlockerID: stableBlockerID(logical, failure.Diagnostics), Decision: decision, Stage: failure.Stage, Step: failure.Step, Reason: failure.Reason, UnknownClass: failure.UnknownClass, NextOperation: failure.NextOperation, BlockedBy: failure.BlockedBy, Diagnostics: failure.Diagnostics}
	}
	return extractionFailureRecord{Logical: logical, Decision: "UNKNOWN", Stage: "apply-extraction", Step: "generic", Reason: "EXTRACTION_FAILED", UnknownClass: "DIRECT_MISSING", NextOperation: "restore-parser-evidence", BlockedBy: []string{}, Diagnostics: []string{}}
}

func stableBlockerID(logical string, diagnostics []string) string {
	for _, diagnostic := range diagnostics {
		if strings.HasPrefix(diagnostic, "declaration=") {
			return logical + "#" + strings.TrimPrefix(diagnostic, "declaration=")
		}
	}
	return ""
}
